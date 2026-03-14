import { useState, useEffect } from 'react'
import { getPurchaseOrders, getVendors } from '../services/api'

const statusStyles = {
  OPEN: 'bg-blue-100 text-blue-700',
  PARTIALLY_MATCHED: 'bg-amber-100 text-amber-700',
  CLOSED: 'bg-gray-100 text-gray-700',
}

export default function PurchaseOrders() {
  const [pos, setPOs] = useState([])
  const [vendorMap, setVendorMap] = useState({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([getPurchaseOrders(), getVendors()])
      .then(([poRes, vendorRes]) => {
        if (poRes.data?.success) setPOs(poRes.data.data || [])
        if (vendorRes.data?.success) {
          const map = {}
          for (const v of vendorRes.data.data || []) map[v.id] = v.name
          setVendorMap(map)
        }
      })
      .catch(() => setError('Failed to load purchase orders'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-gray-500">Loading purchase orders...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-red-600">{error}</p>
      </div>
    )
  }

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">Purchase Orders</h1>
        <p className="text-gray-500 mt-1">Manage purchase orders for invoice matching</p>
      </div>

      {pos.length === 0 ? (
        <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-16 text-center">
          <p className="text-gray-500 font-medium">No purchase orders found</p>
        </div>
      ) : (
        <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase">PO Number</th>
                <th className="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase">Vendor</th>
                <th className="text-right px-6 py-3 text-xs font-medium text-gray-500 uppercase">Total Value</th>
                <th className="text-right px-6 py-3 text-xs font-medium text-gray-500 uppercase">Remaining</th>
                <th className="text-center px-6 py-3 text-xs font-medium text-gray-500 uppercase">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {pos.map((po) => (
                <tr key={po.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-4 text-sm font-medium text-gray-900">{po.po_number}</td>
                  <td className="px-6 py-4 text-sm text-gray-600">{vendorMap[po.vendor_id] || po.vendor_id}</td>
                  <td className="px-6 py-4 text-sm text-gray-900 text-right">
                    ₹{Number(po.total_value).toLocaleString('en-IN')}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-900 text-right">
                    ₹{Number(po.remaining_value).toLocaleString('en-IN')}
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${statusStyles[po.status] || 'bg-gray-100 text-gray-700'}`}>
                      {po.status.replace('_', ' ')}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
