import { API_ORDER_URL, API_PAYMENT_URL } from './config';

export interface OrderProduct {
  id: string;
  name: string;
  price: number;
}

export interface TimelineEvent {
  name: string;
  timestamp: string;
}

export interface Order {
  id: string;
  customer_id: string;
  product: OrderProduct;
  quantity: number;
  total_amount: number;
  status: 'Created' | 'Confirmed' | 'Delivered' | 'Cancelled';
  payment_id?: string;
  created_at: string;
  updated_at: string;
  timeline: TimelineEvent[];
}

export interface CreateOrderRequest {
  customer_id: string;
  product_id: string;
  quantity: number;
}

export const createOrder = async (orderData: CreateOrderRequest): Promise<Order> => {
  const token = localStorage.getItem('token');
  if (!token) {
    throw new Error('No authentication token found');
  }

  const response = await fetch(`${API_ORDER_URL}/orders`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': token,
    },
    body: JSON.stringify(orderData),
  });

  if (!response.ok) {
    throw new Error('Failed to create order');
  }

  const data = await response.json();
  return data.order;
};

export const getOrders = async (): Promise<Order[]> => {
  const token = localStorage.getItem('token');
  if (!token) {
    throw new Error('No authentication token found');
  }

  const response = await fetch(`${API_ORDER_URL}/orders`, {
    headers: {
      'Authorization': token,
    },
  });

  if (!response.ok) {
    throw new Error('Failed to fetch orders');
  }

  return response.json();
};

export const cancelOrder = async (orderId: string): Promise<void> => {
  const token = localStorage.getItem('token');
  if (!token) {
    throw new Error('No authentication token found');
  }

  const response = await fetch(`${API_ORDER_URL}/orders/${orderId}/cancel`, {
    method: 'POST',
    headers: {
      'Authorization': token,
    },
  });

  if (!response.ok) {
    throw new Error('Failed to cancel order');
  }
};

export const initiatePayment = async (orderId: string, amount: number): Promise<void> => {
  const token = localStorage.getItem('token');
  if (!token) {
    throw new Error('No authentication token found');
  }

  const response = await fetch(`${API_PAYMENT_URL}/payments`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': token,
    },
    body: JSON.stringify({ order_id: orderId, amount }),
  });

  if (!response.ok) {
    throw new Error('Failed to initiate payment');
  }
};
