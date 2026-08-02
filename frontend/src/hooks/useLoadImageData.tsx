import { useEffect } from 'react';
import { useAtom } from 'jotai';
import { imageDataAtom, positionsAtom } from './atoms';
import { convertLatLongToVector3, loadImageData } from '../utils';
import worldImage from '../assets/textures/EarthColorMap-4k.jpg';

const DOT_COUNT = 40000;
const RADIUS = 3;

export const useLoadImageData = () => {
  const [imageData, setImageData] = useAtom(imageDataAtom);
  const [, setPositions] = useAtom(positionsAtom);

  useEffect(() => {
    const loadAndProcessImage = async () => {
      try {
        const imageData: any = await loadImageData(worldImage);
        setImageData(imageData);

        const points: number[] = [];
        const phi = Math.PI * (3 - Math.sqrt(5)); // Golden angle in radians

        for (let i = 0; i < DOT_COUNT; i++) {
          const y = 1 - (i / (DOT_COUNT - 1)) * 2; // y goes from 1 to -1
          const radius = Math.sqrt(1 - y * y); // radius at y

          const theta = phi * i; // golden angle increment

          const x = Math.cos(theta) * radius;
          const z = Math.sin(theta) * radius;

          // Convert to latitude and longitude
          const lat = Math.asin(y) * (180 / Math.PI);
          const lon = Math.atan2(z, x) * (180 / Math.PI);

          // Calculate the pixel position on the texture
          const imgX = Math.floor(((lon + 180) / 360) * imageData.width);
          const imgY = Math.floor(((90 - lat) / 180) * imageData.height);
          const idx = (imgY * imageData.width + imgX) * 4;

          // Check if the pixel is part of the land (white in this case)
          const r = imageData.data[idx];
          const g = imageData.data[idx + 1];
          const b = imageData.data[idx + 2];
          const threshold = 30;

          if (!(r < threshold && g < threshold && b < threshold)) {
            const vec3 = convertLatLongToVector3(lat, lon, RADIUS);
            points.push(vec3.x, vec3.y, vec3.z);
          }
        }

        const positionsArray = new Float32Array(points);
        setPositions(positionsArray);
      } catch (error) {
        console.error('Error loading image data:', error);
      }
    };

    if (!imageData) {
      loadAndProcessImage();
    }
  }, [imageData, setImageData, setPositions]);
};
