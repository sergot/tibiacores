'use client';

import { useRouter } from 'next/navigation';

interface List {
  id: string;
  name: string;
  description?: string;
  is_creator: boolean;
  member_count: number;
  created_at: string;
  updated_at: string;
  world: string;
}

interface ListCardProps {
  list: List;
}

export default function ListCard({ list }: ListCardProps) {
  const router = useRouter();

  // Function to format dates
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString();
  };

  // Navigate to list details
  const navigateToList = () => {
    router.push(`/lists/${list.id}`);
  };

  return (
    <div 
      onClick={navigateToList}
      className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 border border-amber-200 dark:border-amber-800 hover:shadow-md transition-shadow cursor-pointer"
    >
      <h3 className="text-lg font-semibold text-amber-800 dark:text-amber-500 mb-2">
        {list.name}
      </h3>
      
      {list.description && (
        <p className="text-gray-700 dark:text-gray-300 mb-4 line-clamp-2">
          {list.description}
        </p>
      )}
      
      <div className="flex flex-wrap gap-2 mb-2">
        {list.is_creator && (
          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-300">
            Owner
          </span>
        )}
        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300">
          {list.member_count} members
        </span>
        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300">
          {list.world}
        </span>
      </div>
      
      <div className="text-xs text-gray-500 dark:text-gray-400 mt-4">
        <div>Created: {formatDate(list.created_at)}</div>
        <div>Last updated: {formatDate(list.updated_at)}</div>
      </div>
    </div>
  );
} 