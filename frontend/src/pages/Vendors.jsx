import { useState, useEffect } from 'react'
import { Building2, X } from 'lucide-react'
import { getVendors, createVendor } from '../services/api'

export default function Vendors() {
  const [vendors, setVendors] = useState([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState({
    name: '', email: '', payment_terms_days: 30,
    early_payment_discount: 0, early_payment_days: 10, criticality_score: 5
  })

  useEffect(() => { fetchVendors() }, [])

  const fetchVendors = async () => {
    try {
      const res = await getVendors()
      setVendors(res.data.data || [])
    } catch (err) {
      console.error('Failed to fetch vendors', err)
    } finally {
      setLoading(false)
    }
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setSaving(true)
    try {
      await createVendor(form)
      setShowModal(false)
      setForm({ name: '', email: '', payment_terms_days: 30, early_payment_discount: 0, early_payment_days: 10, criticality_score: 5 })
      fetchVendors()
    } catch (err) {
      alert('Failed to create vendor: ' + (err.response?.data?.error || err.message))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Vendors</h1>
          <p className="text-gray-500 mt-1">Manage supplier information and payment terms</p>
        </div>
        <button
          onClick={() => setShowModal(true)}
          className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors"
        >
          + Add Vendor
        </button>
      </div>

      {/* Add Vendor Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-md">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-lg font-semibold">Add Vendor</h2>
              <button onClick={() => setShowModal(false)}><X className="w-5 h-5" /></button>
            </div>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Name *</label>
                <input type="text" required value={form.name} onChange={e => setForm({...form, name: e.target.value})}
                  className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-primary-500" />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Email *</label>
                <input type="email" required value={form.email} onChange={e => setForm({...form, email: e.target.value})}
                  className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-primary-500" />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Payment Terms (days)</label>
                  <input type="number" value={form.payment_terms_days} onChange={e => setForm({...form, payment_terms_days: parseInt(e.target.value)})}
                    className="w-full px-3 py-2 border rounded-lg" />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Criticality (1-10)</label>
                  <input type="number" min="1" max="10" value={form.criticality_score} onChange={e => setForm({...form, criticality_score: parseInt(e.target.value)})}
                    className="w-full px-3 py-2 border rounded-lg" />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Early Discount %</label>
                  <input type="number" step="0.01" value={form.early_payment_discount} onChange={e => setForm({...form, early_payment_discount: parseFloat(e.target.value)})}
                    className="w-full px-3 py-2 border rounded-lg" />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Discount Days</label>
                  <input type="number" value={form.early_payment_days} onChange={e => setForm({...form, early_payment_days: parseInt(e.target.value)})}
                    className="w-full px-3 py-2 border rounded-lg" />
                </div>
              </div>
              <button type="submit" disabled={saving}
                className="w-full py-2 bg-primary-600 text-white rounded-lg font-medium hover:bg-primary-700 disabled:opacity-50">
                {saving ? 'Saving...' : 'Add Vendor'}
              </button>
            </form>
          </div>
        </div>
      )}

      {loading ? (
        <div className="text-center py-12 text-gray-500">Loading vendors...</div>
      ) : vendors.length === 0 ? (
        <div className="text-center py-12 text-gray-500">No vendors found. Add one to get started!</div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {vendors.map((vendor) => (
            <div key={vendor.id} className="bg-white rounded-xl border border-gray-200 shadow-sm p-6">
              <div className="flex items-start gap-3 mb-4">
                <div className="p-2 bg-primary-50 rounded-lg">
                  <Building2 className="w-5 h-5 text-primary-600" />
                </div>
                <div>
                  <h3 className="text-sm font-semibold text-gray-900">{vendor.name}</h3>
                  <p className="text-xs text-gray-500">{vendor.email}</p>
                </div>
              </div>
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-gray-500">Payment Terms</span>
                  <span className="text-gray-900 font-medium">Net {vendor.payment_terms_days}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-gray-500">Early Discount</span>
                  <span className="text-gray-900 font-medium">
                    {vendor.early_payment_discount > 0
                      ? `${vendor.early_payment_discount}% / ${vendor.early_payment_days} days`
                      : 'None'}
                  </span>
                </div>
                <div className="flex justify-between text-sm items-center">
                  <span className="text-gray-500">Criticality</span>
                  <div className="flex items-center gap-2">
                    <div className="w-16 h-1.5 bg-gray-200 rounded-full overflow-hidden">
                      <div
                        className={`h-full rounded-full ${
                          vendor.criticality_score >= 8 ? 'bg-red-500' :
                          vendor.criticality_score >= 5 ? 'bg-amber-500' : 'bg-green-500'
                        }`}
                        style={{ width: `${vendor.criticality_score * 10}%` }}
                      />
                    </div>
                    <span className="text-xs text-gray-600">{vendor.criticality_score}/10</span>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
