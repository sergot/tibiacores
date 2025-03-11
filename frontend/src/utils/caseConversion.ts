/**
 * Utility functions for converting between camelCase and snake_case
 */

/**
 * Converts a camelCase string to snake_case
 * @param str The camelCase string to convert
 * @returns The snake_case version of the string
 */
export const camelToSnake = (str: string): string => {
  return str.replace(/[A-Z]/g, letter => `_${letter.toLowerCase()}`);
};

/**
 * Converts a snake_case string to camelCase
 * @param str The snake_case string to convert
 * @returns The camelCase version of the string
 */
export const snakeToCamel = (str: string): string => {
  return str.replace(/_([a-z])/g, (_, letter) => letter.toUpperCase());
};

/**
 * Recursively converts all keys in an object from camelCase to snake_case
 * @param obj The object with camelCase keys
 * @returns A new object with all keys converted to snake_case
 */
export const camelCaseToSnakeCase = <T extends Record<string, any>>(obj: T): Record<string, any> => {
  if (obj === null || typeof obj !== 'object') {
    return obj;
  }

  if (Array.isArray(obj)) {
    return obj.map(item => camelCaseToSnakeCase(item)) as any;
  }

  return Object.keys(obj).reduce((acc, key) => {
    const snakeKey = camelToSnake(key);
    acc[snakeKey] = camelCaseToSnakeCase(obj[key]);
    return acc;
  }, {} as Record<string, any>);
};

/**
 * Recursively converts all keys in an object from snake_case to camelCase
 * @param obj The object with snake_case keys
 * @returns A new object with all keys converted to camelCase
 */
export const snakeCaseToCamelCase = <T extends Record<string, any>>(obj: T): Record<string, any> => {
  if (obj === null || typeof obj !== 'object') {
    return obj;
  }

  if (Array.isArray(obj)) {
    return obj.map(item => snakeCaseToCamelCase(item)) as any;
  }

  return Object.keys(obj).reduce((acc, key) => {
    const camelKey = snakeToCamel(key);
    acc[camelKey] = snakeCaseToCamelCase(obj[key]);
    return acc;
  }, {} as Record<string, any>);
}; 