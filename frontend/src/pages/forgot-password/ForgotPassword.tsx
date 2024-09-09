import React, { useState, FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { forgotPassword } from '../../api/Auth';
import './ForgotPassword.css';

function ForgotPassword() {
  const [email, setEmail] = useState<string>('');
  const [message, setMessage] = useState<string>('');
  const [error, setError] = useState<string>('');
  const navigate = useNavigate();

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError('');
    setMessage('');

    try {
      await forgotPassword(email);
      setMessage('Password reset instructions have been sent to your email.');
      // Redirect to reset-password page after a short delay
      setTimeout(() => {
        navigate('/reset-password', { state: { email } });
      }, 2000);
    } catch (err) {
      setError('Failed to process your request. Please try again.');
    }
  };

  return (
    <div className="forgot-password-page">
      <div className="forgot-password-container">
        <h1>Forgot Password</h1>
        {error && <p className="error-message">{error}</p>}
        {message && <p className="success-message">{message}</p>}
        <form onSubmit={handleSubmit} className="forgot-password-form">
          <div className="form-group">
            <label htmlFor="email">Email:</label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>
          <button type="submit" className="forgot-password-button">Reset Password</button>
        </form>
        <p className="login-link">
          Remember your password? <Link to="/login">Login here</Link>
        </p>
        <p className="reset-password-link">
          Already have a reset token? <Link to="/reset-password">Reset your password here</Link>
        </p>
      </div>
    </div>
  );
}

export default ForgotPassword;
