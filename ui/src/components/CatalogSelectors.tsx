import { useEffect, useState, useRef } from 'react'
import { RefreshCw, AlertCircle, Database, Check, ChevronDown } from 'lucide-react'
import { api } from '../api'
import type { SportType, Facility, Court } from '../types'

interface Props {
  sportType: string
  facilities: string[]
  courts: string[]
  onSportTypeChange: (name: string) => void
  onFacilitiesChange: (names: string[]) => void
  onCourtsChange: (names: string[]) => void
}

export function CatalogSelectors({
  sportType, facilities, courts,
  onSportTypeChange, onFacilitiesChange, onCourtsChange,
}: Props) {
  const [sportTypes, setSportTypes] = useState<SportType[]>([])
  const [facilityList, setFacilityList] = useState<Facility[]>([])
  const [courtList, setCourtList] = useState<Court[]>([])

  const [selectedStId, setSelectedStId] = useState('')
  const [selectedFacIds, setSelectedFacIds] = useState<string[]>([])

  const [loadingSt, setLoadingSt] = useState(true)
  const [loadingFac, setLoadingFac] = useState(false)
  const [loadingCourt, setLoadingCourt] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    api.catalog.sportTypes()
      .then(data => {
        setSportTypes(data)
        if (sportType) {
          const match = data.find(s => s.name.toUpperCase() === sportType.toUpperCase())
          if (match) selectSportTypeById(match, true)
        }
      })
      .catch(e => setError(e.message))
      .finally(() => setLoadingSt(false))
  }, [])

  const selectSportTypeById = (st: SportType, preserveExisting = false) => {
    setSelectedStId(st.id)
    if (!preserveExisting) {
      setSelectedFacIds([])
      setCourtList([])
      onFacilitiesChange([])
      onCourtsChange([])
    }
    onSportTypeChange(st.name)

    setLoadingFac(true)
    api.catalog.facilities(st.id)
      .then(data => {
        setFacilityList(data)
        if (preserveExisting && facilities.length > 0) {
          const matched = data.filter(f =>
            facilities.some(n => n.toUpperCase() === f.name.toUpperCase())
          )
          if (matched.length > 0) {
            const ids = matched.map(f => f.id)
            setSelectedFacIds(ids)
            loadCourtsForFacilities(ids)
          }
        }
      })
      .catch(e => setError(e.message))
      .finally(() => setLoadingFac(false))
  }

  const loadCourtsForFacilities = async (facIds: string[]) => {
    if (facIds.length === 0) { setCourtList([]); return }
    setLoadingCourt(true)
    try {
      const results = await Promise.all(facIds.map(id => api.catalog.courts(id)))
      setCourtList(results.flat())
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Salon yüklenemedi')
    } finally {
      setLoadingCourt(false)
    }
  }

  const toggleFacility = (fac: Facility) => {
    const isSelected = selectedFacIds.includes(fac.id)
    const newIds = isSelected
      ? selectedFacIds.filter(id => id !== fac.id)
      : [...selectedFacIds, fac.id]
    setSelectedFacIds(newIds)
    onFacilitiesChange(facilityList.filter(f => newIds.includes(f.id)).map(f => f.name))
    onCourtsChange([])
    loadCourtsForFacilities(newIds)
  }

  const toggleCourt = (court: Court) => {
    const isSelected = courts.includes(court.name)
    onCourtsChange(isSelected ? courts.filter(n => n !== court.name) : [...courts, court.name])
  }

  if (!loadingSt && sportTypes.length === 0) return (
    <div className="rounded-xl border border-dashed border-slate-600 p-5 text-center">
      <Database className="w-6 h-6 text-slate-600 mx-auto mb-2" />
      <p className="text-slate-400 text-sm mb-1">Katalog boş</p>
      <code className="text-slate-500 text-xs bg-slate-800 px-2 py-1 rounded">
        go run ./cmd/catalog-sync
      </code>
    </div>
  )

  if (error) return (
    <div className="flex items-center gap-2 text-red-400 text-sm">
      <AlertCircle className="w-4 h-4 shrink-0" />{error}
    </div>
  )

  const stLabel = sportType || 'Spor tipi seç...'
  const facLabel = facilities.length === 0
    ? 'Tesis seç...'
    : facilities.length === 1 ? facilities[0] : `${facilities.length} tesis seçildi`
  const courtLabel = courts.length === 0
    ? 'Salon seç... (isteğe bağlı)'
    : courts.length === 1 ? courts[0] : `${courts.length} salon seçildi`

  return (
    <div className="space-y-2">

      {/* Spor Tipi */}
      <DropdownField label="Spor Tipi" valueLabel={stLabel} hasValue={!!sportType} loading={loadingSt}>
        {sportTypes.map(st => (
          <DropdownItem
            key={st.id}
            label={st.name}
            selected={selectedStId === st.id}
            radio
            onClick={() => selectSportTypeById(st)}
          />
        ))}
      </DropdownField>

      {/* Tesisler */}
      <DropdownField
        label="Tesisler"
        valueLabel={facLabel}
        hasValue={facilities.length > 0}
        loading={loadingFac}
        disabled={!selectedStId}
        badge={facilities.length || undefined}
      >
        {facilityList.map(fac => (
          <DropdownItem
            key={fac.id}
            label={fac.name}
            selected={selectedFacIds.includes(fac.id)}
            onClick={() => toggleFacility(fac)}
          />
        ))}
        {!loadingFac && facilityList.length === 0 && (
          <p className="text-slate-500 text-xs px-3 py-2">Bu spor tipine ait tesis bulunamadı.</p>
        )}
      </DropdownField>

      {/* Kortlar / Salonlar */}
      <DropdownField
        label="Kortlar / Salonlar"
        valueLabel={courtLabel}
        hasValue={courts.length > 0}
        loading={loadingCourt}
        disabled={selectedFacIds.length === 0}
        badge={courts.length || undefined}
        hint="Boş bırakılırsa tüm kortlar denenir"
      >
        {courtList.map(court => (
          <DropdownItem
            key={court.id}
            label={court.name}
            selected={courts.includes(court.name)}
            onClick={() => toggleCourt(court)}
          />
        ))}
        {!loadingCourt && courtList.length === 0 && (
          <p className="text-slate-500 text-xs px-3 py-2">Salon verisi bulunamadı.</p>
        )}
      </DropdownField>

    </div>
  )
}

// ─── DropdownField ────────────────────────────────────────────────────────────

function DropdownField({
  label, valueLabel, hasValue, loading, disabled, badge, hint, children,
}: {
  label: string
  valueLabel: string
  hasValue: boolean
  loading?: boolean
  disabled?: boolean
  badge?: number
  hint?: string
  children: React.ReactNode
}) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [open])

  return (
    <div ref={ref} className="relative">
      <div className="flex items-center gap-2 mb-1">
        <span className="text-slate-400 text-xs font-medium">{label}</span>
        {badge ? (
          <span className="px-1.5 py-0.5 bg-emerald-700/60 text-emerald-300 text-[10px] font-bold rounded-full">
            {badge}
          </span>
        ) : null}
        {hint && <span className="text-slate-600 text-[10px]">— {hint}</span>}
      </div>

      <button
        type="button"
        disabled={disabled}
        onClick={() => !disabled && setOpen(o => !o)}
        className={`w-full flex items-center justify-between px-3 py-2.5 rounded-lg border text-sm transition-colors ${
          disabled
            ? 'bg-slate-900/40 border-slate-700 text-slate-600 cursor-not-allowed'
            : open
            ? 'bg-slate-800 border-emerald-600 text-white'
            : hasValue
            ? 'bg-slate-900 border-slate-600 text-white hover:border-slate-500'
            : 'bg-slate-900 border-slate-600 text-slate-500 hover:border-slate-500'
        }`}
      >
        <span className="truncate">{loading ? 'Yükleniyor...' : valueLabel}</span>
        {loading
          ? <RefreshCw className="w-3.5 h-3.5 text-slate-500 animate-spin shrink-0" />
          : <ChevronDown className={`w-3.5 h-3.5 shrink-0 transition-transform text-slate-400 ${open ? 'rotate-180' : ''}`} />
        }
      </button>

      {open && (
        <div className="absolute z-30 left-0 right-0 mt-1 bg-slate-800 border border-slate-700 rounded-xl shadow-2xl shadow-black/40 overflow-hidden">
          <div className="max-h-56 overflow-y-auto py-1">
            {children}
          </div>
        </div>
      )}
    </div>
  )
}

// ─── DropdownItem ─────────────────────────────────────────────────────────────

function DropdownItem({
  label, selected, radio, onClick,
}: {
  label: string
  selected: boolean
  radio?: boolean
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`w-full flex items-center gap-2.5 px-3 py-2 text-left text-sm transition-colors ${
        selected ? 'bg-emerald-900/40 text-white' : 'text-slate-300 hover:bg-slate-700 hover:text-white'
      }`}
    >
      <span className={`shrink-0 w-3.5 h-3.5 flex items-center justify-center border transition-colors ${
        radio ? 'rounded-full' : 'rounded'
      } ${selected ? 'border-emerald-500 bg-emerald-600' : 'border-slate-600'}`}>
        {selected && <Check className="w-2.5 h-2.5 text-white" />}
      </span>
      <span className="truncate">{label}</span>
    </button>
  )
}
