import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Invoices from './pages/Invoices'
import InvoiceDetail from './pages/InvoiceDetail'
import UploadInvoice from './pages/UploadInvoice'
import PurchaseOrders from './pages/PurchaseOrders'
import PaymentSchedule from './pages/PaymentSchedule'
import CashFlow from './pages/CashFlow'
import Vendors from './pages/Vendors'
import Login from './pages/Login'
import { AuthProvider, useAuth } from './context/AuthContext'

function ProtectedRoute({ children }) {
  const { token } = useAuth()
  if (!token) return <Navigate to="/login" />
  return children
}

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/" element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }>
            <Route index element={<Dashboard />} />
            <Route path="invoices" element={<Invoices />} />
            <Route path="invoices/:id" element={<InvoiceDetail />} />
            <Route path="upload" element={<UploadInvoice />} />
            <Route path="purchase-orders" element={<PurchaseOrders />} />
            <Route path="payments" element={<PaymentSchedule />} />
            <Route path="cash-flow" element={<CashFlow />} />
            <Route path="vendors" element={<Vendors />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}

export default App
