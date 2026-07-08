import { useState, useEffect } from 'react';
import type { Product } from '../types';
import api from '../services/api';
import i18n from '../i18n';

export const useProducts = () => {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchProducts = async () => {
      try {
        const { data } = await api.get<Product[]>('/products');
        setProducts(data || []);
      } catch (err: any) {
        setError(err.message || i18n.t('errors.fetchProducts'));
      } finally {
        setLoading(false);
      }
    };

    fetchProducts();
  }, []);

  const getProduct = async (id: number): Promise<Product | undefined> => {
    const { data } = await api.get<Product>(`/products?id=${id}`);
    return data;
  };

  return { products, loading, error, getProduct };
};
