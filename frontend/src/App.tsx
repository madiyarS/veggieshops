import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { Header } from './components/Header';
import { RequireRoles } from './components/RequireRoles';
import { Home } from './pages/Home';
import { Catalog } from './pages/Catalog';
import { Cart } from './pages/Cart';
import { Checkout } from './pages/Checkout';
import { MyOrdersList, MyOrderDetail } from './pages/MyOrders';
import { Login } from './pages/Login';
import { Register } from './pages/Register';
import { AdminStores } from './pages/AdminStores';
import { AdminCategories } from './pages/AdminCategories';
import { AdminProducts } from './pages/AdminProducts';
import { AdminInventory } from './pages/AdminInventory';
import { AdminReports } from './pages/AdminReports';
import { AdminDelivery } from './pages/AdminDelivery';
import { CourierPage } from './pages/CourierPage';

function AppShell() {
  const location = useLocation();
  const path = location.pathname;
  const isPublicAuth = path === '/login' || path === '/register';
  const isStaffApp = path.startsWith('/admin') || path.startsWith('/courier');

  return (
    <div className={isStaffApp ? 'min-h-screen' : 'min-h-screen bg-gray-50'}>
      {!isStaffApp && !isPublicAuth && <Header />}
      <main className={isStaffApp || isPublicAuth ? 'min-h-screen' : 'container mx-auto px-4 py-8'}>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/catalog" element={<Catalog />} />
          <Route
            path="/cart"
            element={
              <RequireRoles roles={['customer']}>
                <Cart />
              </RequireRoles>
            }
          />
          <Route
            path="/checkout"
            element={
              <RequireRoles roles={['customer']}>
                <Checkout />
              </RequireRoles>
            }
          />
          <Route
            path="/orders/:id"
            element={
              <RequireRoles roles={['customer']}>
                <MyOrderDetail />
              </RequireRoles>
            }
          />
          <Route
            path="/orders"
            element={
              <RequireRoles roles={['customer']}>
                <MyOrdersList />
              </RequireRoles>
            }
          />
          <Route path="/track" element={<Navigate to="/orders" replace />} />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />

          <Route path="/admin" element={<Navigate to="/login?next=/admin/stores" replace />} />
          <Route
            path="/admin/stores"
            element={
              <RequireRoles roles={['admin', 'manager']}>
                <AdminStores />
              </RequireRoles>
            }
          />
          <Route
            path="/admin/delivery"
            element={
              <RequireRoles roles={['admin', 'manager']}>
                <AdminDelivery />
              </RequireRoles>
            }
          />
          <Route
            path="/admin/categories"
            element={
              <RequireRoles roles={['admin', 'manager']}>
                <AdminCategories />
              </RequireRoles>
            }
          />
          <Route
            path="/admin/products"
            element={
              <RequireRoles roles={['admin', 'manager']}>
                <AdminProducts />
              </RequireRoles>
            }
          />
          <Route path="/admin/warehouse" element={<Navigate to="/admin/inventory" replace />} />
          <Route
            path="/admin/inventory"
            element={
              <RequireRoles roles={['admin', 'manager']}>
                <AdminInventory />
              </RequireRoles>
            }
          />
          <Route
            path="/admin/reports"
            element={
              <RequireRoles roles={['admin', 'manager']}>
                <AdminReports />
              </RequireRoles>
            }
          />
          <Route path="/admin/*" element={<Navigate to="/login?next=/admin/stores" replace />} />

          <Route
            path="/courier"
            element={
              <RequireRoles roles={['courier']}>
                <CourierPage />
              </RequireRoles>
            }
          />
        </Routes>
      </main>
    </div>
  );
}

function App() {
  return (
    <BrowserRouter>
      <AppShell />
    </BrowserRouter>
  );
}

export default App;
