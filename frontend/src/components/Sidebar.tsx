import React from 'react'
import { FileText, Activity, Layers, Wrench, Settings } from 'lucide-react'
import type { ViewType } from '../context/AppContext'

interface SidebarProps {
  currentView: ViewType
  isWatching: boolean
  onSwitchView: (view: ViewType) => void
  onOpenSettings: () => void
}

export default function Sidebar({ currentView, isWatching, onSwitchView, onOpenSettings }: SidebarProps) {
  const nav = (view: ViewType, label: string, Icon: React.ComponentType<{ size?: number }>, showBadge?: boolean) => (
    <div
      className={`nav-item${currentView === view ? ' active' : ''}`}
      onClick={() => onSwitchView(view)}
      onKeyDown={(e) => e.key === 'Enter' && onSwitchView(view)}
      role="button"
      tabIndex={0}
    >
      <Icon size={18} />
      <span>{label}</span>
      {showBadge && <span className="nav-badge">Live</span>}
    </div>
  )

  return (
    <aside className="sidebar">
      <div className="sidebar-header">
        <div className="app-logo">
          <div className="logo-icon">⧉</div>
          <h2>File Renamer</h2>
        </div>
        <div className={`status-badge${isWatching ? ' watching' : ''}`}>
          <div className="status-indicator" />
          {isWatching ? 'Active' : 'Idle'}
        </div>
      </div>

      <nav className="sidebar-nav">
        <div className="nav-section">
          <h3>Navigation</h3>
          {nav('configuration', 'File Renaming', FileText)}
          {nav('monitoring', 'Live Monitoring', Activity, isWatching)}
          {nav('advanced', 'Advanced Options', Layers)}
          {nav('advancedOperations', 'Advanced Operations', Wrench)}
          <div
            className="nav-item"
            onClick={onOpenSettings}
            onKeyDown={(e) => e.key === 'Enter' && onOpenSettings()}
            role="button"
            tabIndex={0}
          >
            <Settings size={18} />
            <span>Settings</span>
          </div>
        </div>
      </nav>
    </aside>
  )
}
