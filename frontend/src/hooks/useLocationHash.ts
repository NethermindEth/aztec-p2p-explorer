import { useLocation } from 'react-router-dom';

export default function useLocationHash() {
  const hashWithQuery = useLocation().hash.slice(1);
  const queryIndex = hashWithQuery.indexOf('?');
  const hash = queryIndex === -1 ? hashWithQuery : hashWithQuery.slice(0, queryIndex);

  return hash;
}
