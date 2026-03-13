import { Building2 } from 'lucide-react'

export default function Vendors() {
  // TODO: Dev 4 — Replace with API call: getVendors()
  const mockVendors = [
    { name: 'Acme Cloud Solutions', email: 'billing@acmecloud.com', terms: 'Net 30', discount: '2% / 10 days', criticality: 9 },
    { name: 'TechParts India Pvt Ltd', email: 'accounts@techparts.in', terms: 'Net 30', discount: 'None', criticality: 7 },
    { name: 'Global Office Supplies', email: 'finance@globalsupply.com', terms: 'Net 45', discount: '1.5% / 15 days', criticality: 5 },
    { name: 'SecureNet Cybersecurity', email: 'invoices@securenet.io', terms: 'Net 30', discount: '2.5% / 10 days', criticality: 10 },
    { name: 'DataFlow Analytics', email: 'ap@dataflow.co.in', terms: 'Net 60', discount: 'None', criticality: 6 },
  ]

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Vendors</h1>
          <p className="text-gray-500 mt-1">Manage supplier information and payment terms</p>
        </div>
        <button className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors">
          + Add Vendor
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {mockVendors.map((vendor) => (
          <div key={vendor.name} className="bg-white rounded-xl border border-gray-200 shadow-sm p-6">
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
                <span className="text-gray-900 font-medium">{vendor.terms}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-gray-500">Early Discount</span>
                <span className="text-gray-900 font-medium">{vendor.discount}</span>
              </div>
              <div className="flex justify-between text-sm items-center">
                <span className="text-gray-500">Criticality</span>
                <div className="flex items-center gap-2">
                  <div className="w-16 h-1.5 bg-gray-200 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full ${
                        vendor.criticality >= 8 ? 'bg-red-500' :
                        vendor.criticality >= 5 ? 'bg-amber-500' : 'bg-green-500'
                      }`}
                      style={{ width: `${vendor.criticality * 10}%` }}
                    />
                  </div>
                  <span className="text-xs text-gray-600">{vendor.criticality}/10</span>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
