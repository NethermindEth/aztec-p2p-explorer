/**
 * Network configuration based on domain
 */

export type NetworkType = 'aztec-testnet' | 'aztec-mainnet';

// TEMPORARY: Set to true while aztecnodes.xyz is still on testnet
// When ready to make aztecnodes.xyz mainnet, change this to false
const AZTECNODES_IS_TESTNET = false;

/**
 * Detects the current network based on the domain
 */
export function getNetworkFromDomain(): NetworkType {
  if (typeof window === 'undefined') {
    return 'aztec-testnet'; // Default for SSR
  }

  const hostname = window.location.hostname;

  // Check for mainnet subdomain (mainnet.aztecnodes.xyz)
  if (hostname === 'mainnet.aztecnodes.xyz') {
    return 'aztec-mainnet';
  }

  // Check for legacy domain
  if (hostname === 'aztec.nethermind.io') {
    return 'aztec-mainnet';
  }

  // Check for testnet subdomain (testnet.aztecnodes.xyz)
  if (hostname.includes('testnet.')) {
    return 'aztec-testnet';
  }

  // Check for aztecnodes.xyz (temporary testnet, will be mainnet later)
  if (hostname === 'aztecnodes.xyz') {
    return AZTECNODES_IS_TESTNET ? 'aztec-testnet' : 'aztec-mainnet';
  }

  // Check for development environment (aztec-p2p-explorer.dev-nethermind.xyz)
  // Development environment defaults to mainnet
  if (hostname.includes('dev-nethermind.xyz')) {
    return 'aztec-mainnet';
  }

  // Check for local development
  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    // For local dev, check if there's an env variable override
    const envNetwork = import.meta.env.VITE_AZTEC_NETWORK;
    return (envNetwork as NetworkType) || 'aztec-testnet';
  }

  // Default fallback
  return 'aztec-testnet';
}

/**
 * Gets the domain URL for a specific network
 */
export function getDomainForNetwork(network: NetworkType): string {
  const hostname = window.location.hostname;

  // For local development, return current origin (stay on localhost)
  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return window.location.origin;
  }

  // For development environment, redirect to production URLs
  if (hostname.includes('dev-nethermind.xyz')) {
    return network === 'aztec-testnet' ? 'https://testnet.aztecnodes.xyz' : 'https://mainnet.aztecnodes.xyz';
  }

  // For production environments
  if (network === 'aztec-testnet') {
    return 'https://testnet.aztecnodes.xyz';
  } else {
    // Mainnet uses mainnet.aztecnodes.xyz for now
    // When AZTECNODES_IS_TESTNET is set to false, aztecnodes.xyz will become mainnet
    return 'https://mainnet.aztecnodes.xyz';
  }
}

/**
 * Gets the display name for a network
 */
export function getNetworkDisplayName(network: string): string {
  switch (network) {
    case 'aztec-testnet':
      return 'Aztec Testnet';
    case 'aztec-mainnet':
      return 'Aztec Mainnet';
    default:
      return 'Aztec Testnet';
  }
}

/**
 * Navigates to a different network by changing the domain
 */
export function navigateToNetwork(network: NetworkType): void {
  const currentNetwork = getNetworkFromDomain();

  // Don't navigate if we're already on the target network
  if (currentNetwork === network) {
    return;
  }

  const targetDomain = getDomainForNetwork(network);

  // Preserve the current path and search params
  const currentPath = window.location.pathname;
  const currentSearch = window.location.search;
  const currentHash = window.location.hash;

  // Navigate to the new domain with the same path
  window.location.href = `${targetDomain}${currentPath}${currentSearch}${currentHash}`;
}
