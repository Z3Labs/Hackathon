import React from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import Apps from './pages/Apps'
import Machines from './pages/Machines'
import Publish from './pages/Publish'
import Monitor from './pages/Monitor'
import './App.css'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Navigate to="/apps" replace />} />
          <Route path="apps" element={<Apps />} />
          <Route path="machines" element={<Machines />} />
          <Route path="publish" element={<Publish />} />
          <Route path="monitor" element={<Monitor />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
