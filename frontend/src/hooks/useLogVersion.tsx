import { useEffect } from 'react';

const useLogVersion = () => {
  useEffect(() => {
    const version = import.meta.env.VITE_APP_VERSION;
    if (version) {
      console.log(`App Version: ${version}`);
    }
  }, []);
};

export default useLogVersion;
