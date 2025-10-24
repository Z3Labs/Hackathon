import React from 'react'
import { Link, Outlet, useLocation } from 'react-router-dom'
import './Layout.css'

const Layout: React.FC = () => {
  const location = useLocation()

  const menuItems = [
    { path: '/publish', label: '发布' },
    { path: '/monitor', label: '监控' }
  ]

  return (
    <div className="layout">
      <aside className="sidebar">
        <div className="logo">
          <h2>Hackathon</h2>
        </div>
        <nav className="menu">
          {menuItems.map(item => (
            <Link
              key={item.path}
              to={item.path}
              className={`menu-item ${location.pathname === item.path ? 'active' : ''}`}
            >
              {item.label}
            </Link>
          ))}
        </nav>
      </aside>
      <main className="content">
        <Outlet />
      </main>
    </div>
  )
}

export default Layout
