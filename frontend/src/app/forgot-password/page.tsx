'use client';

import { useState } from 'react';
import Link from 'next/link';
import { authApi } from '@/services/authApi';

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    
    if (!email) {
      setError('Email is required');
      return;
    }
    
    setIsSubmitting(true);
    
    try {
      // This would typically call the password reset API
      // For now, we'll just show a success message
      setSuccess(true);
    } catch (err: any) {
      console.error('Password reset error:', err);
      setError(err.message || 'Failed to send password reset email. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };
  
  return (
    <div className="max-w-md mx-auto">
      <h1 className="text-2xl font-bold text-amber-800 dark:text-amber-400 mb-6">Reset Your Password</h1>
      
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
        {success ? (
          <div className="text-center">
            <div className="flex justify-center mb-6">
              <div className="rounded-full bg-green-100 p-3">
                <svg className="h-8 w-8 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              </div>
            </div>
            
            <h2 className="text-xl font-semibold text-gray-800 dark:text-white mb-4">Check Your Email</h2>
            
            <p className="text-gray-600 dark:text-gray-300 mb-6">
              If an account exists with the email <span className="font-medium">{email}</span>, we've sent instructions to reset your password.
            </p>
            
            <Link
              href="/login"
              className="inline-block px-4 py-2 bg-amber-500 hover:bg-amber-600 text-white rounded-md transition-colors"
            >
              Return to Login
            </Link>
          </div>
        ) : (
          <>
            <p className="text-gray-600 dark:text-gray-300 mb-6">
              Enter your email address and we'll send you instructions to reset your password.
            </p>
            
            {error && (
              <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mb-4" role="alert">
                <span className="block sm:inline">{error}</span>
              </div>
            )}
            
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label htmlFor="email" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Email
                </label>
                <input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-amber-500 dark:bg-gray-700 dark:text-white"
                  placeholder="your@email.com"
                  required
                />
              </div>
              
              <button
                type="submit"
                disabled={isSubmitting}
                className="w-full px-4 py-2 bg-amber-500 hover:bg-amber-600 text-white rounded-md transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isSubmitting ? 'Sending...' : 'Send Reset Instructions'}
              </button>
            </form>
            
            <div className="mt-6 text-center">
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Remember your password?{' '}
                <Link href="/login" className="text-amber-600 hover:text-amber-500 dark:text-amber-400 dark:hover:text-amber-300">
                  Log in
                </Link>
              </p>
            </div>
          </>
        )}
      </div>
    </div>
  );
} 