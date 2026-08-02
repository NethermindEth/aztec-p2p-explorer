import { ReactNode } from 'react';
import MapView from './views/MapView/MapView';
import ListView from './views/ListView/ListView';
import NotFoundPage from './views/NotFoundPage/NotFoundPage';

type AppRoute = {
  path: string;
  element: ReactNode;
  name?: string;
};

const appRoutes: AppRoute[] = [
  {
    path: '/',
    element: <MapView />,
  },
  {
    path: '/explore',
    element: <ListView />,
  },
  {
    path: '*',
    element: <NotFoundPage />,
  },
];

export default appRoutes;
