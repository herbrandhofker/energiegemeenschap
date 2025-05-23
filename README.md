# Tibber Loader

Een applicatie voor het laden van energieverbruik en -productie data van de Tibber API.

## Installatie

1. Installeer de benodigde dependencies:
```bash
go mod download
```

## Configuratie

Configureer je Tibber API gegevens via environment variables:
```bash
export TIBBER_API_TOKEN=<REDACTED>
export DATABASE_URL=<REDACTED>
```

## Database Structuur

De applicatie gebruikt PostgreSQL met de volgende structuur:

```
energiegemeenschap (database)
└── tibber (schema)
    ├── tibber_tokens
    │   ├── id (SERIAL PRIMARY KEY)
    │   ├── token (VARCHAR(255))
    │   ├── active (BOOLEAN DEFAULT true)
    │   ├── created_at (TIMESTAMP)
    │   └── updated_at (TIMESTAMP)
    │
    ├── homes
    │   ├── id (VARCHAR(255) PRIMARY KEY)
    │   ├── name (VARCHAR(255))
    │   ├── address (VARCHAR(255))
    │   ├── has_production_capability (BOOLEAN)
    │   ├── created_at (TIMESTAMP)
    │   └── updated_at (TIMESTAMP)
    │
    ├── prices
    │   ├── id (SERIAL PRIMARY KEY)
    │   ├── home_id (VARCHAR(255) REFERENCES homes)
    │   ├── timestamp (TIMESTAMP)
    │   ├── total_price (DECIMAL)
    │   ├── energy_price (DECIMAL)
    │   ├── tax_price (DECIMAL)
    │   ├── currency (VARCHAR)
    │   ├── created_at (TIMESTAMP)
    │   └── updated_at (TIMESTAMP)
    │
    ├── consumption
    │   ├── id (SERIAL PRIMARY KEY)
    │   ├── home_id (VARCHAR(255) REFERENCES homes)
    │   ├── timestamp (TIMESTAMP)
    │   ├── consumption (DECIMAL)
    │   ├── unit (VARCHAR)
    │   ├── created_at (TIMESTAMP)
    │   └── updated_at (TIMESTAMP)
    │
    ├── production
    │   ├── id (SERIAL PRIMARY KEY)
    │   ├── home_id (VARCHAR(255) REFERENCES homes)
    │   ├── timestamp (TIMESTAMP)
    │   ├── production (DECIMAL)
    │   ├── unit (VARCHAR)
    │   ├── created_at (TIMESTAMP)
    │   └── updated_at (TIMESTAMP)
    │
    └── real_time_measurements
        ├── id (SERIAL PRIMARY KEY)
        ├── home_id (VARCHAR(255) REFERENCES homes)
        ├── timestamp (TIMESTAMP)
        ├── power (DECIMAL)
        ├── power_production (DECIMAL)
        ├── accumulated_consumption (DECIMAL)
        ├── accumulated_production (DECIMAL)
        ├── created_at (TIMESTAMP)
        └── updated_at (TIMESTAMP)
```

## Trigger Implementatie

De applicatie gebruikt PostgreSQL triggers om wijzigingen in de `tibber_tokens` tabel te monitoren. Hier is de implementatie:

```sql
-- Functie die wordt aangeroepen bij wijzigingen in tibber_tokens
CREATE OR REPLACE FUNCTION tibber.notify_token_changes()
RETURNS TRIGGER AS $$
BEGIN
    -- Stuur een NOTIFY met de actie en token ID
    PERFORM pg_notify(
        'token_changes',
        json_build_object(
            'action', TG_OP,
            'token_id', NEW.id,
            'active', NEW.active
        )::text
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger voor INSERT en UPDATE
CREATE TRIGGER token_changes_trigger
    AFTER INSERT OR UPDATE ON tibber.tibber_tokens
    FOR EACH ROW
    EXECUTE FUNCTION tibber.notify_token_changes();
```

De applicatie luistert naar deze notificaties met:

```go
// Voorbeeld Go code voor het luisteren naar notificaties
func listenForTokenChanges(db *sql.DB) {
    listener := pq.NewListener(connStr, 10*time.Second, time.Minute, nil)
    err := listener.Listen("token_changes")
    if err != nil {
        log.Fatal(err)
    }

    for {
        select {
        case notification := <-listener.Notify:
            var event struct {
                Action  string `json:"action"`
                TokenID int    `json:"token_id"`
                Active  bool   `json:"active"`
            }
            json.Unmarshal([]byte(notification.Extra), &event)
            
            switch event.Action {
            case "INSERT":
                if event.Active {
                    startDataCollection(event.TokenID)
                }
            case "UPDATE":
                if event.Active {
                    startDataCollection(event.TokenID)
                } else {
                    stopDataCollection(event.TokenID)
                }
            }
        case <-time.After(90 * time.Second):
            // Check connection
            if err := listener.Ping(); err != nil {
                log.Printf("Listener connection lost: %v", err)
            }
        }
    }
}
```

## Proces Beschrijving

Het data laden proces werkt als volgt:

1. **Token Monitoring**
   - De applicatie monitort de `tibber_tokens` tabel via PostgreSQL triggers
   - Bij elke INSERT of UPDATE wordt een notificatie gestuurd
   - Alleen tokens met `active = true` worden gebruikt voor data laden

2. **Home Data Laden**
   - Wanneer een nieuw actief token wordt toegevoegd:
     - Verbinding maken met Tibber API met het nieuwe token
     - Ophalen van alle homes gekoppeld aan het token
     - Homes worden opgeslagen in de `homes` tabel

3. **Metingen Laden**
   - Voor elk opgehaald home:
     - Start het laden van real-time metingen
     - Start het laden van historische consumptie data
     - Start het laden van historische productie data (indien beschikbaar)
     - Start het laden van prijsinformatie

4. **Proces Controle**
   - Het laden kan worden gestopt door `active = false` te zetten voor een token
   - Het laden wordt automatisch hervat wanneer `active = true` wordt gezet
   - Bij het deactiveren van een token worden alle lopende metingen voor dat token gestopt

## Gebruik

De applicatie kan worden uitgevoerd met:

```bash
go run cmd/collector/main.go
```

Dit zal:
1. Starten met monitoren van de `tibber_tokens` tabel via PostgreSQL triggers
2. Automatisch data laden voor alle actieve tokens
3. De data opslaan in de PostgreSQL database