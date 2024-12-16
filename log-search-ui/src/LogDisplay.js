import React from 'react';

const LogDisplay = ({ logs }) => {
  const formatTimestamp = (timestamp) => {
    try {
      const date = new Date(timestamp);
      return date.toLocaleString();
    } catch (error) {
      return timestamp;
    }
  };

  const formatMetadata = (metadata) => {
    if (!metadata) return 'N/A';
    try {
      if (typeof metadata === 'string') {
        metadata = JSON.parse(metadata);
      }
      return (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
          {Object.entries(metadata).map(([key, value], index) => (
            <div key={index} style={{ display: 'flex', gap: '0.5rem' }}>
              <span style={{ color: '#4b5563', minWidth: '100px' }}>{key}:</span>
              <span style={{ color: '#1f2937' }}>{String(value)}</span>
            </div>
          ))}
        </div>
      );
    } catch (error) {
      return String(metadata);
    }
  };

  return (
    <div style={{ overflowX: 'auto' }}>
      <table style={{
        width: '100%',
        borderCollapse: 'collapse',
        marginTop: '1rem',
        backgroundColor: 'white',
        boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1)'
      }}>
        <thead>
          <tr style={{
            backgroundColor: '#f3f4f6',
            borderBottom: '2px solid #e5e7eb'
          }}>
            <th style={{
              padding: '0.75rem',
              textAlign: 'left',
              fontWeight: '600',
              width: '200px'
            }}>
              Timestamp
            </th>
            <th style={{
              padding: '0.75rem',
              textAlign: 'left',
              fontWeight: '600'
            }}>
              Raw Log
            </th>
            <th style={{
              padding: '0.75rem',
              textAlign: 'left',
              fontWeight: '600',
              width: '250px'
            }}>
              Metadata
            </th>
          </tr>
        </thead>
        <tbody>
          {logs.length > 0 ? (
            logs.map((log, index) => (
              <tr 
                key={index}
                style={{
                  borderBottom: '1px solid #e5e7eb',
                  backgroundColor: index % 2 === 0 ? '#ffffff' : '#f9fafb'
                }}
              >
                <td style={{
                  padding: '0.75rem',
                  fontSize: '0.875rem',
                  whiteSpace: 'nowrap',
                  verticalAlign: 'top'
                }}>
                  {formatTimestamp(log.timestamp)}
                </td>
                <td style={{
                  padding: '0.75rem',
                  fontSize: '0.875rem',
                  verticalAlign: 'top'
                }}>
                  <div style={{
                    maxWidth: '400px',
                    overflowX: 'auto',
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-word'
                  }}>
                    {log.raw_log || log.message || 'N/A'}
                  </div>
                </td>
                <td style={{
                  padding: '0.75rem',
                  fontSize: '0.875rem',
                  verticalAlign: 'top'
                }}>
                  <div style={{
                    maxWidth: '300px',
                    overflowX: 'auto',
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-word'
                  }}>
                    {formatMetadata(log.data.message)}
                  </div>
                </td>
              </tr>
            ))
          ) : (
            <tr>
              <td 
                colSpan="3" 
                style={{
                  padding: '1rem',
                  textAlign: 'center',
                  color: '#6b7280'
                }}
              >
                No logs found
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
};

export default LogDisplay;