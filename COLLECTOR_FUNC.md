# Tibber Data Collector Functionaliteit

## Overzicht
De Tibber Data Collector bestaat uit twee hoofdcomponenten die onafhankelijk van elkaar werken:

1. Real-time data verzameling
2. Historische data laden

## 1. Real-time Collector (`real_time.go`)
De real-time collector verzamelt continue metingen van Tibber huizen met productievermogen.

### Functionaliteit
- Maakt WebSocket verbinding met Tibber API
- Haalt real-time metingen op voor huizen met productievermogen
- Slaat metingen op in de `real_time_measurements` tabel
- Draait continu in een loop met updates elke 5 minuten

### Configuratie
- Vereist `DATABASE_URL` in .env bestand
- Vereist `TIBBER_API_TOKEN` in .env bestand
- Optioneel `TIBBER_HOUSE_ID` voor specifieke huizen
- De naamgeving van de Tibber API is leidend, maar kan aangepast worden naar behoefte.

## 2. Historische Data (`historical.go`)
De historische data loader haalt historische gegevens op bij het opstarten van het programma.

### Functionaliteit
Laadt voor alle huizen met productievermogen:
- Prijzen (huidige, vandaag en morgen) in de `prices` tabel
  deze worden elk uur ververst. Rond 1 uur 's middags komt er nieuwe data beschikbaar
- Consumptie data van de laatste 30 dagen in de `consumption` tabel
- Productie data van de laatste 30 dagen in de `production` tabel
  Consumptie en productie dienen elke nacht rond 3 uur  aangevuld te wqorden met nieuwe data


### Configuratie
- Vereist `DATABASE_URL` in .env bestand
- Vereist `TIBBER_API_TOKEN` in .env bestand

## Database Tabellen

### real_time_measurements
Bevat real-time metingen:
- Timestamp
- Vermogen (power)
- Productievermogen (power_production)
- Min/max/gemiddeld vermogen
- Geaccumuleerd verbruik/productie
- Laatste meterstanden
- Spanning per fase
- Stroom per fase

### prices
Bevat prijsinformatie per uur:
- Home ID
- Datum
- Uur van de dag
- Totaalprijs
- Energieprijs
- Belasting
- Valuta
- Prijsniveau

### consumption
Bevat dagelijkse verbruiksdata:
- Home ID
- Van datum
- Tot datum
- Verbruik
- Kosten
- Valuta

### production
Bevat dagelijkse productiedata:
- Home ID
- Van datum
- Tot datum
- Productie
- Opbrengst
- Valuta 