export default function PurchaseOrders() {
  // TODO: Dev 4 — Replace with API call: getPurchaseOrders()
  const mockPOs = [
    { po_number: 'PO-2026-100', vendor: 'Acme Cloud Solutions', total: 50000, remaining: 50000, status: 'OPEN' },
    { po_number: 'PO-2026-101', vendor: 'TechParts India Pvt Ltd', total: 125000, remaining: 125000, status: 'OPEN' },
    { po_number: 'PO-2026-102', vendor: 'Global Office Supplies', total: 35000, remaining: 35000, status: 'OPEN' },
    { po_number: 'PO-2026-103', vendor: 'SecureNet Cybersecurity', total: 200000, remaining: 200000, status: 'OPEN' },
    { po_number: 'PO-2026-104', vendor: 'DataFlow Analytics', total: 75000, remaining: 75000, status: 'OPEN' },
    { po_number: 'PO-2026-105', vendor: 'Acme Cloud Solutions', total: 80000, remaining: 30000, status: 'PARTIALLY_MATCHED' },
    { po_number: 'PO-2026-106', vendor: 'TechParts India Pvt Ltd', total: 45000, remaining: 0, status: 'CLOSED' },
  ]

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">Purchase Orders</h1>
        <p className="text-gray-500 mt-1">Manage purchase orders for invoice matching</p>
      </div>

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
            {mockPOs.map((po) => (
              <tr key={po.po_number} className="hover:bg-gray-50 transition-colors">
                <td className="px-6 py-4 text-sm font-medium text-gray-900">{po.po_number}</td>
                <td className="px-6 py-4 text-sm text-gray-600">{po.vendor}</td>
                <td className="px-6 py-4 text-sm text-gray-900 text-right">₹{po.total.toLocaleString('en-IN')}</td>
                <td className="px-6 py-4 text-sm text-gray-900 text-right">₹{po.remaining.toLocaleString('en-IN')}</td>
                <td className="px-6 py-4 text-center">
                  <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${
                    po.status === 'OPEN' ? 'bg-blue-100 text-blue-700' :
                    po.status === 'PARTIALLY_MATCHED' ? 'bg-amber-100 text-amber-700' :
                    'bg-gray-100 text-gray-700'
                  }`}>
                    {po.status.replace('_', ' ')}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
