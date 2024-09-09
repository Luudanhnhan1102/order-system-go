import React, { useState, FormEvent, useEffect } from 'react';
import { useNavigate, Link, useLocation } from 'react-router-dom';
import { resetPassword } from '../../api/Auth';
import './ResetPassword.css';

function ResetPassword() {
  const [email, setEmail] = useState<string>('');
  const [resetToken, setResetToken] = useState<string>('');
  const [newPassword, setNewPassword] = useState<string>('');
  const [confirmPassword, setConfirmPassword] = useState<string>('');
  const [message, setMessage] = useState<string>('');
  const [error, setError] = useState<string>('');
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    const state = location.state as { email?: string };
    if (state && state.email) {
      setEmail(state.email);
    }
  }, [location]);

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError('');
    setMessage('');

    if (newPassword !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    try {
      await resetPassword(email, resetToken, newPassword);
      setMessage('Password has been reset successfully.');
      setTimeout(() => navigate('/login'), 3000);
    } catch (err) {
      setError('Failed to reset password. Please try again.');
    }
  };

  return (
    <div className="reset-password-page">
      <div className="reset-password-container">
        <h1>Reset Password</h1>
        {error && <p className="error-message">{error}</p>}
        {message && <p className="success-message">{message}</p>}
        <form onSubmit={handleSubmit} className="reset-password-form">
          <div className="form-group">
            <label htmlFor="email">Email:</label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              readOnly={!!location.state}
            />
          </div>
          <div className="form-group">
            <label htmlFor="resetToken">Reset Token:</label>
            <input
              type="text"
              id="resetToken"
              value={resetToken}
              onChange={(e) => setResetToken(e.target.value)}
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="newPassword">New Password:</label>
            <input
              type="password"
              id="newPassword"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="confirmPassword">Confirm New Password:</label>
            <input
              type="password"
              id="confirmPassword"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
            />
          </div>
          <button type="submit" className="reset-password-button">Reset Password</button>
        </form>
        <p className="login-link">
          Remember your password? <Link to="/login">Login here</Link>
        </p>
      </div>
    </div>
  );
}

export default ResetPassword;
