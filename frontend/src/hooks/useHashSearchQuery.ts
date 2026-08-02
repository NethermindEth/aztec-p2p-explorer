import { useLocation } from 'react-router-dom';

export default function useHashSearchQuery() {
  const hashWithQuery = useLocation().hash;
  const queryIndex = hashWithQuery.indexOf('?');
  const query = hashWithQuery.slice(queryIndex + 1);

  return new URLSearchParams(query);
}
