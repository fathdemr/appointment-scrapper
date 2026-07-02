import { Link, useLocation } from 'react-router-dom'
import { CalendarCheck, LayoutDashboard, PlusCircle } from 'lucide-react'

export function Navbar() {
  const { pathname } = useLocation()

  return (
    <nav className="bg-slate-900 border-b border-slate-700/50 sticky top-0 z-40">
      <div className="max-w-6xl mx-auto px-4 sm:px-6">
        <div className="flex items-center justify-between h-16">
          <Link to="/" className="flex items-center gap-2.5 group">
            <div className="w-8 h-8 rounded-lg bg-emerald-600 flex items-center justify-center">
              <CalendarCheck className="w-4 h-4 text-white" />
            </div>
            <span className="text-white font-semibold text-sm sm:text-base hidden xs:block">
              Appointment Scrapper
            </span>
          </Link>

          <div className="flex items-center gap-1">
            <Link
              to="/"
              className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                pathname === '/'
                  ? 'bg-slate-700 text-white'
                  : 'text-slate-400 hover:text-white hover:bg-slate-800'
              }`}
            >
              <LayoutDashboard className="w-4 h-4" />
              <span className="hidden sm:block">Dashboard</span>
            </Link>
            <Link
              to="/jobs/new"
              className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium bg-emerald-600 hover:bg-emerald-500 text-white transition-colors"
            >
              <PlusCircle className="w-4 h-4" />
              <span className="hidden sm:block">Yeni Job</span>
            </Link>
          </div>
        </div>
      </div>
    </nav>
  )
}
