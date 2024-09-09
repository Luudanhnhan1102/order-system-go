import React, { useState, useEffect } from 'react';
import { fetchProducts } from '../../api/Product';
import { createOrder } from '../../api/Order';
import { useAuth } from '../../contexts/AuthContext'; // Adjust the import path as needed
import './Products.css';

interface Product {
  id: string;
  name: string;
  price: number;
  stocks: number;
}

const Products: React.FC = () => {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [showDialog, setShowDialog] = useState<boolean>(false);
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null);
  const [quantity, setQuantity] = useState<number>(1);

  // Get the current user's email from the auth context
  const { userEmail } = useAuth();

  useEffect(() => {
    const loadProducts = async () => {
      try {
        const fetchedProducts = await fetchProducts();
        setProducts(fetchedProducts);
        setLoading(false);
      } catch (err) {
        setError('Failed to fetch products. Please try again later.');
        setLoading(false);
      }
    };

    loadProducts();
  }, []);

  const handleBuy = (product: Product) => {
    setSelectedProduct(product);
    setShowDialog(true);
    setQuantity(1); // Reset quantity when opening dialog
  };

  const handleClose = () => {
    setShowDialog(false);
    setSelectedProduct(null);
  };

  const handleConfirmOrder = async () => {
    if (selectedProduct && userEmail) {
      try {
        const orderData = {
          customer_id: userEmail,
          product_id: selectedProduct.id,
          quantity: quantity
        };
        await createOrder(orderData);
        console.log(`Ordered ${quantity} of ${selectedProduct.name}`);
        
        // Update local state
        const updatedProducts = products.map(p => 
          p.id === selectedProduct.id ? { ...p, stocks: p.stocks - quantity } : p
        );
        setProducts(updatedProducts);
        handleClose();
      } catch (error) {
        console.error('Failed to create order:', error);
        setError('Failed to create order. Please try again.');
      }
    } else if (!userEmail) {
      setError('You must be logged in to place an order.');
    }
  };

  const handleQuantityChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newQuantity = Math.max(1, Math.min(parseInt(e.target.value) || 0, selectedProduct?.stocks || 0));
    setQuantity(newQuantity);
  };

  if (loading) {
    return <div>Loading products...</div>;
  }

  if (error) {
    return <div>{error}</div>;
  }

  return (
    <div className="products-container">
      <h2 className="products-header">Products</h2>
      <div className="products-list">
        {products.map((product) => (
          <div key={product.id} className="product-item">
            <div className="product-id">ID: {product.id}</div>
            <div className="product-name">{product.name}</div>
            <div className="product-price">Price: ${product.price.toFixed(2)}</div>
            <div className="product-stock">Stock: {product.stocks}</div>
            <button className="buy-button" onClick={() => handleBuy(product)}>
              Buy Now
            </button>
          </div>
        ))}
      </div>

      {showDialog && selectedProduct && (
        <div className="dialog-overlay">
          <div className="dialog">
            <h3>Place Order</h3>
            {userEmail ? (
              <>
                <p>Product: {selectedProduct.name}</p>
                <p>Price: ${selectedProduct.price.toFixed(2)}</p>
                <p>Stock left: {selectedProduct.stocks}</p>
                <p>Ordering as: {userEmail}</p>
                <label htmlFor="quantity">Quantity:</label>
                <input
                  id="quantity"
                  type="number"
                  value={quantity}
                  onChange={handleQuantityChange}
                  min={1}
                  max={selectedProduct.stocks}
                />
                <p>Subtotal: ${(quantity * selectedProduct.price).toFixed(2)}</p>
                {error && <p className="error-message">{error}</p>}
                <div className="dialog-actions">
                  <button onClick={handleClose}>Cancel</button>
                  <button onClick={handleConfirmOrder}>Confirm Order</button>
                </div>
              </>
            ) : (
              <p>Please log in to place an order.</p>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default Products;
