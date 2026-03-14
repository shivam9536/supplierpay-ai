import axios from 'axios'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

const api = axios.create({
  baseURL: `${API_URL}/api/v1`,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Add auth token to every request
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Handle 401 responses
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// ── Auth ────────────────────────────────────
export const login = (email, password) =>
  api.post('/auth/login', { email, password })

// ── Invoices ────────────────────────────────
export const getInvoices = (params) =>
  api.get('/invoices', { params })

export const getInvoice = (id) =>
  api.get(`/invoices/${id}`)

export const getInvoiceAuditLog = (id) =>
  api.get(`/invoices/${id}/audit-log`)

export const uploadInvoice = (formData) =>
  api.post('/invoices/upload-sync', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })

export const uploadInvoiceJSON = (payload) =>
  api.post('/invoices/upload-json', payload)

export const reprocessInvoice = (id) =>
  api.post(`/invoices/${id}/reprocess`)

// ── Vendors ─────────────────────────────────
export const getVendors = () =>
  api.get('/vendors')

export const getVendor = (id) =>
  api.get(`/vendors/${id}`)

export const createVendor = (vendor) =>
  api.post('/vendors', vendor)

// ── Purchase Orders ─────────────────────────
export const getPurchaseOrders = (params) =>
  api.get('/purchase-orders', { params })

export const getPurchaseOrder = (id) =>
  api.get(`/purchase-orders/${id}`)

// ── Payments ────────────────────────────────
export const getPaymentSchedule = () =>
  api.get('/payments/schedule')

export const triggerPaymentRun = () =>
  api.post('/payments/run')

export const getPaymentRuns = () =>
  api.get('/payments/runs')

// ── Forecast ────────────────────────────────
export const getCashFlowForecast = () =>
  api.get('/forecast')

// ── SSE (Server-Sent Events) ────────────────
export const subscribeToInvoiceUpdates = (invoiceId, onEvent) => {
  const eventSource = new EventSource(
    `${API_URL}/api/v1/events/invoices/${invoiceId}`
  )

  eventSource.onmessage = (event) => {
    const data = JSON.parse(event.data)
    onEvent(data)
  }

  eventSource.onerror = () => {
    eventSource.close()
  }

  return eventSource // Caller should close when done
}

export default api
