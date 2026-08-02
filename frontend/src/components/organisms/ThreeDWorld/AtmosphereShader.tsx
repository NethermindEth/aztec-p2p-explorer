import React, { useRef } from 'react';
import * as THREE from 'three';

const AtmosphereShader = {
  uniforms: {},
  vertexShader: `
    varying vec3 vNormal;
    void main() {
      vNormal = normalize(normalMatrix * normal);
      gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
    }
  `,
  fragmentShader: `
    varying vec3 vNormal;
    void main() {
      float intensity = pow(0.8 - dot(vNormal, vec3(0, 0, 1.0)), 12.0);
      vec3 atmosphereColor = vec3(0.765, 0.773, 0.875); // #C3C5DF in RGB
      gl_FragColor = vec4(atmosphereColor, 1.0) * intensity;
    }
  `,
};

const Atmosphere: React.FC = () => {
  const materialRef = useRef<THREE.ShaderMaterial>(null!);
  const { uniforms, vertexShader, fragmentShader } = AtmosphereShader;

  return (
    <>
      <mesh scale={[1.13, 1.13, 1.13]}>
        <sphereGeometry args={[3.54, 32, 32]} />
        <shaderMaterial
          ref={materialRef}
          uniforms={THREE.UniformsUtils.clone(uniforms)}
          vertexShader={vertexShader}
          fragmentShader={fragmentShader}
          side={THREE.BackSide}
          blending={THREE.AdditiveBlending}
          transparent
          depthWrite={false}
        />
      </mesh>
    </>
  );
};

export default Atmosphere;
