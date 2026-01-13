# go-doc-service
# Example of using options with input field

``` Code ```
import { createFileRoute } from '@tanstack/react-router';
import { useState, useMemo, useRef, useEffect } from 'react'


export const Route = createFileRoute('/')({
  component: App,
})



function App() {
  
const [languages, setLanguages] = useState(["Javascript","Python","TypeScript"])
  const [query, setQuery] = useState("")
  const [isOpen, setIsOpen] = useState(false)
  const inputRef = useRef<HTMLInputElement | null>(null)

  // Filter suggestions based on current input
  const suggestions = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return languages
    return languages.filter(l => l.toLowerCase().includes(q))
  }, [languages, query])

  // Close suggestions on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (!inputRef.current) return
      const target = e.target as Node
      // If the click is outside the input and suggestion container, close
      const container = inputRef.current.closest('.suggestion-container')
      if (container && !container.contains(target)) setIsOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  const handleSelectSuggestion = (value: string) => {
    setQuery(value)     // fill input with the clicked suggestion
    setIsOpen(false)    // hide the dropdown
  }

  const handleInputFocus = () => setIsOpen(true)

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setQuery(e.target.value)
    setIsOpen(true)
  }

  return (
    <div className="min-h-screen flex items-start justify-center p-8">
      <div className="w-full max-w-md suggestion-container">
        <label className="block mb-2 text-sm font-medium text-gray-700">
          Language
        </label>

        {/* Input */}
        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={handleInputChange}
          onFocus={handleInputFocus}
          placeholder="Type to search…"
          className="w-full rounded-lg border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
        />

        {/* Suggestions dropdown */}
        {isOpen && suggestions.length > 0 && (
          <ul className="mt-2 max-h-44 overflow-auto rounded-lg border border-gray-200 bg-white shadow-sm">
            {suggestions.map((s, idx) => (
              <li
                key={`${s}-${idx}`}
                onClick={() => handleSelectSuggestion(s)}
                className="cursor-pointer px-3 py-2 hover:bg-blue-50"
              >
                {s}
              </li>
            ))}
          </ul>
        )}

        {/* Empty state */}
        {isOpen && suggestions.length === 0 && (
          <div className="mt-2 rounded-lg border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-500">
            No matches
          </div>
        )}

        {/* Optional: quick tags below */}
        <div className="mt-4 flex flex-wrap gap-2">
          {languages.map((lang) => (
            <button
              key={lang}
              type="button"
              onClick={() => handleSelectSuggestion(lang)}
              className="rounded-full border border-gray-300 px-3 py-1 text-sm hover:bg-gray-100"
            >
              {lang}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
```
