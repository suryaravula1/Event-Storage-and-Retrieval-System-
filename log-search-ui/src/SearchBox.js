import React, { useState, useEffect, useCallback } from 'react';
import LogDisplay from './LogDisplay';
import DatePicker from 'react-datepicker';
import 'react-datepicker/dist/react-datepicker.css';

const CustomInput = React.forwardRef(({ value, onClick }, ref) => (
  <input
    type="text"
    value={value}
    onClick={onClick}
    ref={ref}
    style={{
      padding: '0.5rem',
      border: '1px solid #ccc',
      borderRadius: '4px',
    }}
  />
));

const SearchBox = () => {
  const [logs, setLogs] = useState([]);
  const [startTime, setStartTime] = useState(new Date(new Date().setSeconds(0, 0)));
  const [endTime, setEndTime] = useState(new Date(new Date().setSeconds(0, 0)));
  const [page, setPage] = useState(1);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [offset, setOffset] = useState(0);
  const [regex, setRegex] = useState('');
  const [text, setText] = useState('');

  useEffect(() => {
    const now = new Date();
    const fiveMinutesAgo = new Date(now.getTime() - 5 * 60000);
    setStartTime(new Date(fiveMinutesAgo.setSeconds(0, 0)).toISOString());
    setEndTime(new Date(now.setSeconds(0, 0)).toISOString());
    console.log('startTime:', fiveMinutesAgo.toISOString().slice(0, -1));
    
  }, []);

  const validateTimeRange = () => {
    if (!startTime && !endTime) {
      setError('Please select at least one time range');
      return false;
    }

    if (startTime && endTime && new Date(startTime) > new Date(endTime)) {
      setError('Start time cannot be later than end time');
      return false;
    }

    return true;
  };

  const fetchLogs = async (reset = false) => {
    console.log('fetchLogs called');
    if (!validateTimeRange()) {
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      // Fetch the UUID
      console.log('fetching uuid', startTime, endTime, regex, text);
      
      const uuidResponse = await fetch(
        `http://localhost:3000/search?startTime=${startTime}&endTime=${endTime}&regex=${regex}&text=${text}`,
        {
          method: 'GET',
          headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json',
          },
          mode: 'cors',
        }
      );
      console.log('uuidResponse:', uuidResponse);
      

      if (!uuidResponse.ok) {
        throw new Error(`HTTP error! status: ${uuidResponse.status}`);
      }

      const uuidData = await uuidResponse.json();
      console.log('uuidData:', uuidData);
      
      const uuid = uuidData.uuid;
      console.log('uuid:', uuid);

      // Fetch the logs using the UUID and offset
      const logsResponse = await fetch(
        `http://localhost:3000/search/${uuid}/${reset ? 0 : offset}?regex=${regex}&text=${text}`,
        {
          method: 'GET',
          headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json',
          },
          mode: 'cors',
        }
      );

      if (!logsResponse.ok) {
        throw new Error(`HTTP error! status: ${logsResponse.status}`);
      }

      const logsData = await logsResponse.json();
      console.log('logsData:', logsData);
      

      if (logsData) {
        if (reset) {
          setLogs(Array.isArray(logsData.data) ? logsData.data : []);
          setOffset(logsData.offset || 0);

          setPage(2);
        } else {
          setLogs(prevLogs => [
            ...prevLogs,
            ...(Array.isArray(logsData.data) ? logsData.data : [])
          ]);
          setOffset(logsData.offset || 0);
          setPage(prev => prev + 1);
        }

        
      }
    } catch (error) {
      console.error('Error fetching logs:', error);
      setError(error.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleSearch = () => {
    // Reset the logs when doing a new search
    setLogs([]);
    setPage(1);
    fetchLogs(true);
  };

  const handleClear = () => {
    setStartTime(startTime);
    setEndTime(endTime);
    setLogs([]);
    setError(null);
    setPage(1);
    setRegex('');
    setText('');
  };

  const handleScroll = useCallback(() => {
    const logDisplay = document.getElementById('log-display');
    
    
    
    if (logDisplay) {
      const { scrollTop, scrollHeight, clientHeight } = logDisplay;
      
      if (scrollTop + clientHeight >= scrollHeight - 5 && !isLoading) {
        console.log('Reached bottom of scroll');
        fetchLogs(false);
      }
    }
  }, [isLoading]);

  useEffect(() => {
    const logDisplay = document.getElementById('log-display');
    if (logDisplay) {
      logDisplay.addEventListener('scroll', handleScroll);
      return () => logDisplay.removeEventListener('scroll', handleScroll);
    }
  }, [handleScroll]);

  const handleStartTimeChange = (date) => {
    setStartTime(new Date(date.setSeconds(0, 0)).toISOString());
  };

  const handleEndTimeChange = (date) => {
    setEndTime(new Date(date.setSeconds(0, 0)).toISOString());
  };

  return (
    <div style={{ padding: '1rem', maxWidth: '60%', margin: '0 auto' }}>
      <h2 style={{ fontSize: '1.5rem', fontWeight: 'bold', marginBottom: '1rem' }}>Log Search</h2>
      
      <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
        <div style={{ 
          display: 'flex', 
          gap: '1rem', 
          flexWrap: 'wrap',
          padding: '1rem',
          backgroundColor: '#f9fafb',
          borderRadius: '8px',
          border: '1px solid #e5e7eb'
        }}>
          <label style={{ display: 'flex', flexDirection: 'column' }}>
            <span style={{ marginBottom: '0.5rem' }}>Time Range:</span>
            <select
              onChange={(e) => {
                const now = new Date();
                let newStartTime;
                switch (e.target.value) {
                  case '5m':
                    newStartTime = new Date(now.getTime() - 5 * 60000);
                    break;
                  case '30m':
                    newStartTime = new Date(now.getTime() - 30 * 60000);
                    break;
                  case '1h':
                    newStartTime = new Date(now.getTime() - 60 * 60000);
                    break;
                  case '12h':
                    newStartTime = new Date(now.getTime() - 12 * 60 * 60000);
                    break;
                  case '1d':
                    newStartTime = new Date(now.getTime() - 24 * 60 * 60000);
                    break;
                  case '3d':
                    newStartTime = new Date(now.getTime() - 3 * 24 * 60 * 60000);
                    break;
                  case '7d':
                    newStartTime = new Date(now.getTime() - 7 * 24 * 60 * 60000);
                    break;
                  default:
                    newStartTime = now;
                }
                setStartTime(new Date(newStartTime.setSeconds(0, 0)).toISOString());
                setEndTime(new Date(now.setSeconds(0, 0)).toISOString());
              }}
              style={{
                padding: '0.5rem',
                border: '1px solid #ccc',
                borderRadius: '4px'
              }}
            >
              <option value="5m">Last 5 minutes</option>
              <option value="30m">Last 30 minutes</option>
              <option value="1h">Last 1 hour</option>
              <option value="12h">Last 12 hours</option>
              <option value="1d">Last 1 day</option>
              <option value="3d">Last 3 days</option>
              <option value="7d">Last 7 days</option>
            </select>
          </label>
          
          <label style={{ display: 'flex', flexDirection: 'column' }}>
            <span style={{ marginBottom: '0.5rem' }}>Start Time:</span>
            <DatePicker
              selected={new Date(startTime)}
              onChange={handleStartTimeChange}
              showTimeSelect
              dateFormat="Pp"
              customInput={<CustomInput />}
            />
          </label>
          
          <label style={{ display: 'flex', flexDirection: 'column' }}>
            <span style={{ marginBottom: '0.5rem' }}>End Time:</span>
            <DatePicker
              selected={new Date(endTime)}
              onChange={handleEndTimeChange}
              showTimeSelect
              dateFormat="Pp"
              customInput={<CustomInput />}
            />
          </label>
          
          <label style={{ display: 'flex', flexDirection: 'column' }}>
            <span style={{ marginBottom: '0.5rem' }}>Regex:</span>
            <input
              type="text"
              value={regex}
              onChange={(e) => setRegex(e.target.value)}
              style={{ 
                padding: '0.5rem',
                border: '1px solid #ccc',
                borderRadius: '4px'
              }}
            />
          </label>
          
          <label style={{ display: 'flex', flexDirection: 'column' }}>
            <span style={{ marginBottom: '0.5rem' }}>Text:</span>
            <input
              type="text"
              value={text}
              onChange={(e) => setText(e.target.value)}
              style={{ 
                padding: '0.5rem',
                border: '1px solid #ccc',
                borderRadius: '4px'
              }}
            />
          </label>
          
          <div style={{ 
            display: 'flex', 
            gap: '0.5rem', 
            alignSelf: 'flex-end'
          }}>
            <button
              onClick={handleSearch}
              disabled={isLoading}
              style={{
                padding: '0.5rem 1rem',
                backgroundColor: isLoading ? '#93c5fd' : '#3b82f6',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: isLoading ? 'not-allowed' : 'pointer',
              }}
            >
              {isLoading ? 'Loading...' : 'Fetch Logs'}
            </button>

            <button
              onClick={handleClear}
              disabled={isLoading}
              style={{
                padding: '0.5rem 1rem',
                backgroundColor: '#dc2626',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: isLoading ? 'not-allowed' : 'pointer',
              }}
            >
              Clear
            </button>
          </div>
        </div>

        {error && (
          <div style={{
            padding: '1rem',
            backgroundColor: '#fee2e2',
            border: '1px solid #ef4444',
            borderRadius: '4px',
            color: '#dc2626'
          }}>
            {error}
          </div>
        )}

        {logs.length > 0 && (
          <>
            <div style={{
              padding: '0.5rem',
              backgroundColor: '#f0fdf4',
              border: '1px solid #86efac',
              borderRadius: '4px',
              color: '#166534'
            }}>
              Showing logs from{' '}
              {startTime ? new Date(startTime).toISOString : 'the beginning'}{' '}
              to{' '}
              {endTime ? new Date(endTime).toISOString : 'now'}
            </div>
            
            <div id="log-display" style={{ overflowY: 'auto', maxHeight: '400px', border: '1px solid #e5e7eb', borderRadius: '8px', padding: '1rem', backgroundColor: '#f9fafb' }}>
              
              <LogDisplay logs={logs} />
            </div>

            {isLoading && (
              <div style={{ textAlign: 'center', padding: '1rem' }}>
                <span>Loading...</span>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
};

export default SearchBox;