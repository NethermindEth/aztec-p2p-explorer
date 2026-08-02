import * as THREE from 'three';

export const convertLatLongToVector3 = (lat: number, lon: number, radius = 3) => {
  const phi = (90 - lat) * (Math.PI / 180);
  const theta = (lon + 180) * (Math.PI / 180);
  return new THREE.Vector3(
    -(radius * Math.sin(phi) * Math.cos(theta)),
    radius * Math.cos(phi),
    radius * Math.sin(phi) * Math.sin(theta)
  );
};

export const applyRotation = (obj: any, lat: number, lon: number) => {
  const position = convertLatLongToVector3(lat, lon);
  obj.position.copy(position);

  obj.lookAt(new THREE.Vector3(0, 0, 0));
  obj.rotateY(Math.PI);

  const rot = lon;
  obj.rotateZ(((-rot * Math.PI) / 10) * 10);
};

export const applyTextRotation = (text: any) => {
  text.lookAt(new THREE.Vector3(0, 0, 0));
  text.rotateY(Math.PI);
};

export const createArc = (
  startLat: number,
  startLon: number,
  endLat: number,
  endLon: number,
  radius = 3,
  segments = 100
) => {
  const startVec = convertLatLongToVector3(startLat, startLon, radius);
  const endVec = convertLatLongToVector3(endLat, endLon, radius);

  const arcPoints = [];
  for (let i = 0; i <= segments; i++) {
    const t = i / segments;
    const arcPoint = new THREE.Vector3()
      .lerpVectors(startVec, endVec, t)
      .normalize()
      .multiplyScalar(radius + Math.sin(Math.PI * t) * 0.8); // Adjust the multiplier for a smooth arc
    arcPoints.push(arcPoint);
  }
  return arcPoints;
};

export const loadAndParseData = async (url: string) => {
  const response = await fetch(url);
  const text = await response.text();
  const data: any = [];
  const settings = { data };
  let max: number = -Infinity;
  let min: number = Infinity;

  text.split('\n').forEach((line) => {
    const parts = line.trim().split(/\s+/);
    if (parts.length === 2) {
      settings[parts[0] as keyof typeof settings] = parseFloat(parts[1]);
    } else if (parts.length > 2) {
      const values = parts.map((v) => {
        const value = parseFloat(v);
        if (value === (settings as any).NODATA_value) {
          return undefined;
        }
        max = Math.max(max === undefined ? value : max, value);
        min = Math.min(min === undefined ? value : min, value);
        return value;
      });
      data.push(values);
    }
  });

  return Object.assign(settings, { min, max });
};

export const loadImageData = (imageSrc: string) => {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.crossOrigin = 'Anonymous';
    img.onload = () => {
      const canvas = document.createElement('canvas');
      canvas.width = img.width;
      canvas.height = img.height;
      const ctx = canvas.getContext('2d');
      ctx!.drawImage(img, 0, 0);
      try {
        const imageData = ctx!.getImageData(0, 0, img.width, img.height);
        console.log('Image loaded and processed:', imageData);
        resolve(imageData);
      } catch (error) {
        console.error('Error getting image data:', error);
        reject(error);
      }
    };
    img.onerror = (err) => {
      console.error('Error loading image:', err);
      reject(err);
    };
    img.src = imageSrc;
  });
};

export const formatDateTime = (dateString: string) => {
  const date = new Date(dateString);

  // Extract the components of the date
  const day = String(date.getDate()).padStart(2, '0');
  const month = String(date.getMonth() + 1).padStart(2, '0'); // Months are 0-indexed
  const year = date.getFullYear();

  // Extract the components of the time
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');

  // Format the date and time as needed
  return `${day}-${month}-${year} - ${hours}:${minutes}`;
};

export const capitalizeFirstLetter = (string: string) => {
  return string.charAt(0).toUpperCase() + string.slice(1);
};

export const getBackgroundStyle = (percentage: number) => {
  return {
    background: `linear-gradient(to right, rgba(0, 196, 216, 0.05) ${percentage}%, transparent ${percentage}%)`,
  };
};

export const isSafari = () => {
  const ua = window.navigator.userAgent;
  return ua.includes('Safari') && !ua.includes('Chrome');
};

export const truncateNodeID = (id: string, startChars: number = 12, endChars: number = 4) => {
  if (id.length <= startChars + endChars) return id;
  return `${id.slice(0, startChars)}...${id.slice(-endChars)}`;
};

// Compute visible characters for a middle-truncated ID based on a target pixel width.
// Always reserves 3 chars for dots and 4 trailing characters; the remainder goes to the start.
export const computeVisibleCharsForWidth = (targetPx: number, averageCharPx = 8): number => {
  if (!isFinite(targetPx) || targetPx <= 0) return 18; // sensible fallback
  const dots = 3;
  const trailing = 4;
  const budget = Math.floor(targetPx / Math.max(averageCharPx, 6));
  return Math.max(trailing + 4 + dots, budget + dots); // return total visible incl. dots
};

export const truncateMiddleWithBudget = (value: string, totalVisible: number) => {
  if (!isFinite(totalVisible) || value.length <= totalVisible) return value;
  const trailing = 4;
  const dots = 3;
  const startLen = Math.max(4, totalVisible - trailing - dots);
  return `${value.slice(0, startLen)}...${value.slice(-trailing)}`;
};
