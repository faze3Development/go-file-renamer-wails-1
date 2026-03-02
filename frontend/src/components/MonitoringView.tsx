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
                <div className="placeholder-icon">◊</div>
                <p>Activity logs will appear here when you start watching...</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
