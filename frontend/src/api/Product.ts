interface Product {
  id: string;
  name: string;
  price: number;
  stocks: number;
}

export const fetchProducts = async (): Promise<Product[]> => {
  const response = await fetch('http://localhost:8080/products');
  if (!response.ok) {
    throw new Error('Failed to fetch products');
  }
  return response.json();
};
