import React, { useState, useEffect } from 'react';
import { getOrders, cancelOrder, initiatePayment, Order } from '../../api/Order';
import OrderTimeline from './OrderTimeline';
import './Orders.css';

const Orders: React.FC = () => {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchOrders();
  }, []);

  const fetchOrders = async () => {
    try {
      const fetchedOrders = await getOrders();
      setOrders(fetchedOrders);
      setLoading(false);
    } catch (err) {
      setError('Failed to fetch orders. Please try again later.');
      setLoading(false);
    }
  };

  const handleCancelOrder = async (orderId: string) => {
    try {
      await cancelOrder(orderId);
      // Refresh the orders list after cancellation
      fetchOrders();
    } catch (err) {
      setError('Failed to cancel order. Please try again later.');
    }
  };

  const handlePayOrder = async (orderId: string, amount: number) => {
    try {
      await initiatePayment(orderId, amount);
      // Refresh the orders list after payment initiation
      fetchOrders();
    } catch (err) {
      setError('Failed to initiate payment. Please try again later.');
    }
  };

  if (loading) {
    return <div className="orders-container">Loading orders...</div>;
  }

  if (error) {
    return <div className="orders-container">{error}</div>;
  }

  return (
    <div className="orders-container">
      <h2 className="orders-header">My Orders</h2>
      {orders.length === 0 ? (
        <p>You have no orders yet.</p>
      ) : (
        <div className="orders-list">
          {orders.map((order) => (
            <div key={order.id} className="order-item">
              <div className="order-details">
                <p>Order ID: {order.id}</p>
                <p>Product: {order.product.name}</p>
                <p>Price: ${order.product.price.toFixed(2)}</p>
                <p>Quantity: {order.quantity}</p>
                <p>Total Amount: ${order.total_amount.toFixed(2)}</p>
                <p>Status: {order.status}</p>
                {order.status === 'Created' && (
                  <div className="order-actions">
                    <button onClick={() => handleCancelOrder(order.id)}>Cancel Order</button>
                    <button onClick={() => handlePayOrder(order.id, order.total_amount)}>Pay Now</button>
                  </div>
                )}
              </div>
              <div className="order-timeline">
                <OrderTimeline
                  status={order.status}
                  timeline={order.timeline}
                />
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default Orders;
