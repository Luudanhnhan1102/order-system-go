import React from 'react';
import { useAuth } from '../../contexts/AuthContext';
import './Header.css';

const Header: React.FC = () => {
  const { isAuthenticated, userEmail, logout } = useAuth();

  return (
    <header className="header">
      <h1>User Order Page</h1>
      {isAuthenticated && (
        <div className="user-info">
          <span>{userEmail}</span>
          <button onClick={logout}>Logout</button>
        </div>
      )}
    </header>
  );
};

export default Header;
