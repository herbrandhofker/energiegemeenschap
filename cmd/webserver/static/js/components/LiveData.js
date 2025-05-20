import {
  CategoryScale,
  Chart as ChartJS,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  Title,
  Tooltip
} from 'chart.js';
import React, { useEffect, useState } from 'react';
import { Line } from 'react-chartjs-2';
import { useDispatch, useSelector } from 'react-redux';

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
);

const LiveData = ({ homeId }) => {
  const dispatch = useDispatch();
  const measurements = useSelector(state => state.measurements[homeId] || []);
  const [eventSource, setEventSource] = useState(null);

  useEffect(() => {
    // Create SSE connection
    const source = new EventSource(`/api/live-data/${homeId}`);
    setEventSource(source);

    source.onmessage = (event) => {
      const data = JSON.parse(event.data);
      dispatch({ type: 'ADD_MEASUREMENT', payload: { homeId, measurement: data } });
    };

    source.onerror = (error) => {
      console.error('SSE Error:', error);
      source.close();
    };

    return () => {
      if (source) {
        source.close();
      }
    };
  }, [homeId, dispatch]);

  const chartData = {
    labels: measurements.map(m => new Date(m.timestamp).toLocaleTimeString()),
    datasets: [
      {
        label: 'Power Consumption (W)',
        data: measurements.map(m => m.power),
        borderColor: 'rgb(255, 99, 132)',
        tension: 0.1
      },
      {
        label: 'Power Production (W)',
        data: measurements.map(m => m.powerProduction),
        borderColor: 'rgb(75, 192, 192)',
        tension: 0.1
      }
    ]
  };

  const options = {
    responsive: true,
    plugins: {
      legend: {
        position: 'top',
      },
      title: {
        display: true,
        text: 'Live Power Data'
      }
    },
    scales: {
      y: {
        beginAtZero: true
      }
    }
  };

  const latestMeasurement = measurements[measurements.length - 1];

  return (
    <div className="live-data-container">
      <div className="chart-container">
        <Line data={chartData} options={options} />
      </div>
      {latestMeasurement && (
        <div className="current-values">
          <h3>Current Values</h3>
          <p>Power Consumption: {latestMeasurement.power.toFixed(2)} W</p>
          <p>Power Production: {latestMeasurement.powerProduction.toFixed(2)} W</p>
          <p>Total Consumption: {latestMeasurement.accumulatedConsumption.toFixed(2)} kWh</p>
          <p>Total Production: {latestMeasurement.accumulatedProduction.toFixed(2)} kWh</p>
        </div>
      )}
    </div>
  );
};

export default LiveData; 