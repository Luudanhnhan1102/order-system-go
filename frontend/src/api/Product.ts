import { API_ORDER_URL } from './config';

interface Product {
  id: string;
  name: string;
  price: number;
  stocks: number;
}

export const fetchProducts = async (): Promise<Product[]> => {
  const response = await fetch(`${API_ORDER_URL}/products`);
  if (!response.ok) {
    throw new Error('Failed to fetch products');
  }
  return response.json();
};
