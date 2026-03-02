import React, { useRef, useEffect } from 'react'
import { BookText } from 'lucide-react'
import type { StatsPayload, LogEntryPayload } from '../types/events'

interface Props {
  stats: StatsPayload
  logs: LogEntryPayload[]
}

export default function MonitoringView({ stats, logs }: Props) {
  const logContainerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const el = logContainerRef.current
    if (el) {
      setTimeout(() => { el.scrollTop = el.scrollHeight }, 50)
    }
  }, [logs])

  const getSeverityClass = (severity: string | undefined) => {
    switch ((severity ?? 'info').toUpperCase()) {
      case 'ERROR': return 'log-error'
      case 'WARN': return 'log-warn'
      case 'DEBUG': return 'log-debug'
      default: return 'log-info'
    }
  }

  return (
    <div className="monitoring-section">
      {/* Stats Panel */}
      <div className="stats-section">
        <div className="stats-header">
          <h3>Live Statistics</h3>
          <p>Real-time monitoring data</p>
        </div>
        <div className="stats-grid">
          <div className="stat-card">
            <div className="stat-icon">◉</div>
            <div className="stat-content">
              <div className="stat-value">{stats.scanned ?? 0}</div>
              <div className="stat-label">Files Scanned</div>
            </div>
          </div>
          <div className="stat-card success">
            <div className="stat-icon">◆</div>
            <div className="stat-content">
              <div className="stat-value">{stats.renamed ?? 0}</div>
              <div className="stat-label">Files Renamed</div>
            </div>
          </div>
          <div className="stat-card warning">
            <div className="stat-icon">◇</div>
            <div className="stat-content">
              <div className="stat-value">{stats.skipped ?? 0}</div>
              <div className="stat-label">Files Skipped</div>
            </div>
          </div>
          <div className="stat-card error">
            <div className="stat-icon">◈</div>
            <div className="stat-content">
              <div className="stat-value">{stats.errors ?? 0}</div>
              <div className="stat-label">Errors</div>
            </div>
          </div>
        </div>
      </div>

      {/* Log Viewer */}
      <div className="log-section">
        <div className="log-header">
          <h3><BookText size={20} />Activity Log</h3>
          <p>Real-time operation details</p>
        </div>
        <div className="log-viewer">
          <div className="log-container" ref={logContainerRef}>
            {logs.map((log, i) => (
              <div key={log.id ?? i} className={`log-entry ${getSeverityClass(log.severity)}`}>
                <span className="log-timestamp">
                  {new Date(log.timestamp ?? '').toLocaleTimeString()}
                </span>
                <span className="log-level-badge">
                  {(log.severity ?? 'info').toUpperCase()}
                </span>
                <span className="log-message">{log.message}</span>
              </div>
            ))}
            {logs.length === 0 && (
              <div className="log-placeholder">
                <div className="empty-state-icon" aria-hidden="true">
                  <svg width="48" height="48" viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <rect x="6" y="10" width="36" height="28" rx="4" stroke="currentColor" strokeWidth="2" strokeDasharray="4 3"/>
                    <line x1="12" y1="19" x2="24" y2="19" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                    <line x1="12" y1="25" x2="30" y2="25" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                    <line x1="12" y1="31" x2="20" y2="31" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                    <circle cx="38" cy="12" r="6" fill="var(--accent-primary)" opacity="0.15" stroke="var(--accent-primary)" strokeWidth="1.5"/>
                    <path d="M35.5 12h5M38 9.5v5" stroke="var(--accent-primary)" strokeWidth="1.5" strokeLinecap="round"/>
                  </svg>
                </div>
                <p className="empty-state-title">No activity yet</p>
                <p className="empty-state-hint">Start watching a directory to see rename operations and events here in real time.</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
