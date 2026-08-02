import React, { useEffect, useRef } from 'react';
import AnimatedLine from './AnimatedLine';
import { useTransition } from '@react-spring/three';

type LatLon = {
  lat: number;
  lon: number;
};

type Connection = {
  from: LatLon;
  to: LatLon;
  id: string;
};

type ConnectionsProps = {
  connections: Connection[];
  setConnections: React.Dispatch<React.SetStateAction<Connection[]>>;
};

const Connections: React.FC<ConnectionsProps> = ({ connections, setConnections }) => {
  const timeouts = useRef<ReturnType<typeof setTimeout>[]>([]);
  const animationDuration = 2500;

  const transitions = useTransition(connections, {
    keys: (item) => item.from.lat + '-' + item.from.lon + '-' + item.to.lat + '-' + item.to.lon,
    from: { opacity: 0, animatedValue: 0 },
    enter: { opacity: 1, animatedValue: 1 },
    leave: { opacity: 0 },
    config: { duration: animationDuration },
    onRest: (_result, _ctrl, item) => {
      setConnections((prev) => prev.filter((conn) => conn.id !== item.id));
    },
  });

  useEffect(() => {
    // Clear any existing timeouts
    timeouts.current.forEach((timeout) => clearTimeout(timeout));
    timeouts.current = [];

    // Set a timeout to remove the connection after the animation duration
    connections.forEach((connection) => {
      const timeout = setTimeout(() => {
        setConnections((prev) => prev.filter((conn) => conn.id !== connection.id));
      }, animationDuration - 100);
      timeouts.current.push(timeout);
    });

    return () => {
      // Clear timeouts on unmount
      timeouts.current.forEach((timeout) => clearTimeout(timeout));
    };
  }, [connections, setConnections]);

  return (
    <>
      {transitions((props, item) => (
        <AnimatedLine
          key={item.from.lat + '-' + item.from.lon + '-' + item.to.lat + '-' + item.to.lon}
          from={item.from}
          to={item.to}
          // @ts-ignore
          animatedProps={props}
        />
      ))}
    </>
  );
};

export default Connections;
