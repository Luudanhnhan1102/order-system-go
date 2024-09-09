import React from 'react';
import { NavLink } from 'react-router-dom';
import './Sidebar.css';
import { FaBox, FaShoppingCart } from 'react-icons/fa';

const Sidebar: React.FC = () => {
  return (
    <nav className="sidebar">
      <ul>
        <li>
          <NavLink to="/products" className={({ isActive }) => isActive ? "active" : ""}>
            <FaBox className="sidebar-icon" />
            Products
          </NavLink>
        </li>
        <li>
          <NavLink to="/orders" className={({ isActive }) => isActive ? "active" : ""}>
            <FaShoppingCart className="sidebar-icon" />
            Orders
          </NavLink>
        </li>
      </ul>
    </nav>
  );
};

export default Sidebar;
