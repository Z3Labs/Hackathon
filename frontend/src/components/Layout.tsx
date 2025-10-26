import React, { useState, useEffect } from 'react'
import { Link, Outlet, useLocation } from 'react-router-dom'
import './Layout.css'

const Layout: React.FC = () => {
  const location = useLocation()
  const [theme, setTheme] = useState<'light' | 'dark'>(() => {
    const savedTheme = localStorage.getItem('theme')
    return (savedTheme as 'light' | 'dark') || 'light'
  })

  const menuItems = [
    { path: '/publish', label: '发布' },
    { path: '/apps', label: '应用管理' },
    { path: '/machines', label: '机器管理' }
  ]

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
    localStorage.setItem('theme', theme)
  }, [theme])

  const toggleTheme = () => {
    setTheme(prevTheme => prevTheme === 'light' ? 'dark' : 'light')
  }

  return (
    <div className="layout">
      <aside className="sidebar">
        <div className="logo">
          <h1>Z³Labs</h1>
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
        <header className="content-header">
          <button className="theme-toggle" onClick={toggleTheme} title={theme === 'light' ? '切换到暗黑模式' : '切换到明亮模式'}>
            {theme === 'light' ? '🌙' : '☀️'}
          </button>
        </header>
        <div className="content-body">
          <Outlet />
        </div>
      </main>
    </div>
  )
}

export default Layout
