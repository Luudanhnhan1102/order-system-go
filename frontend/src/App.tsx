import React from 'react';
import { BrowserRouter as Router, Route, Routes, Navigate, useLocation } from 'react-router-dom';
import Login from './pages/login/Login';
import Register from './pages/register/Register'; 
import ForgotPassword from './pages/forgot-password';
import ResetPassword from './pages/reset-password';
import Products from './pages/products/Products';
import Orders from './pages/orders/Orders';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import Header from './components/header/Header';
import Sidebar from './components/sidebar/Sidebar';
import './App.css';

const ProtectedRoute: React.FC<{ element: React.ReactElement }> = ({ element }) => {
  const { isAuthenticated } = useAuth();
  const location = useLocation();

  if (isAuthenticated) {
    return element;
  } else {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }
};

function AppContent() {
  const { isAuthenticated } = useAuth();
  const location = useLocation();
  const isAuthPage = location.pathname === '/login' || location.pathname === '/register';

  return (
    <div className="App">
      {!isAuthPage && isAuthenticated && <Header />}
      <div className="content-wrapper">
        {!isAuthPage && isAuthenticated && <Sidebar />}
        <main className={!isAuthPage && isAuthenticated ? 'with-sidebar' : ''}>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/forgot-password" element={<ForgotPassword />} />
            <Route path="/reset-password" element={<ResetPassword />} /> {/* Add this line */}
            <Route path="/products" element={<ProtectedRoute element={<Products />} />} />
            <Route path="/orders" element={<ProtectedRoute element={<Orders />} />} />
            <Route path="/" element={<Navigate to="/products" replace />} />
          </Routes>
        </main>
      </div>
    </div>
  );
}

function App() {
  return (
    <AuthProvider>
        <Router>
          <AppContent />
        </Router>
    </AuthProvider>
  );
}

export default App;
