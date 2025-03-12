'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';

export default function VerifyEmailSuccessPage() {
  const [countdown, setCountdown] = useState(5);
  const router = useRouter();
  
  useEffect(() => {
    // Redirect to login page after 5 seconds
    const timer = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          clearInterval(timer);
          router.push('/login');
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
    
    return () => clearInterval(timer);
  }, [router]);
  
  return (
    <div className="max-w-md mx-auto text-center">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-8 mb-6">
        <div className="flex justify-center mb-6">
          <div className="rounded-full bg-green-100 p-3">
            <svg className="h-12 w-12 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
        </div>
        
        <h1 className="text-2xl font-bold text-gray-800 dark:text-white mb-4">Email Verified!</h1>
        
        <p className="text-gray-600 dark:text-gray-300 mb-6">
          Your email has been successfully verified. You can now log in to your account.
        </p>
        
        <p className="text-gray-500 dark:text-gray-400 text-sm mb-6">
          Redirecting to login page in {countdown} seconds...
        </p>
        
        <Link
          href="/login"
          className="inline-block w-full px-4 py-2 bg-amber-500 hover:bg-amber-600 text-white rounded-md transition-colors text-center"
        >
          Log In Now
        </Link>
      </div>
    </div>
  );
} 