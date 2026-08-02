# Aztec P2P Explorer

Aztec P2P Explorer is a tool for monitoring and observing the Aztec peer-to-peer network. It visualizes the network nodes on a 3D globe and provides various statistics about the network.

## Features

- 3D visualization of network nodes on a globe.
- Statistics on hosting/network, client types, geographic locations, operating systems, client versions, and network health.
- Built with React

## Prerequisites

Before you begin, ensure you have met the following requirements:

- Node.js (>= 20.x)
- pnpm (>= 9.x)

## Installation

To install and set up the project, follow these steps:

1. Clone the repository:

2. Navigate to the project directory:

   ```sh
   cd aztec-p2p-explorer/frontend
   ```

3. Install the dependencies:
   ```sh
   pnpm install
   ```

## Running the Project

To run the complete project with backend and frontend:

1. Make sure the maxmind files are in the `maxmind-db` directory in the root folder.

2. Grant permission to the script by running:

   ```sh
   chmod +x start-local-dev.sh
   ```

3. To start the backend and frontend:
   ```sh
   pnpm local-dev
   ```
   or
   ```sh
   ./start-local-dev.sh
   ```

To run only the frontend, start the development server:

```sh
pnpm dev
```

5. Open your browser and navigate to:
   ```
   http://localhost:5173
   ```

You should see the Aztec P2P Explorer application running.

## Environment Variables

Create a file named .env, and add the following:

`VITE_BASE_URL=http://localhost:8080`

## Code Quality

This project uses ESLint and Prettier to maintain code quality. You can run the following commands to lint and format your code:

To run ESLint:

```sh
pnpm lint
```

To run Prettier:

```sh
pnpm format
```

## Project Structure

The main project structure is as follows:

```
frontend/
├── .husky/
│   ├── pre-commit
├── .storybook/
├── node_modules/
├── public/
│   ├── favicons/
│   └── ...
├── src/
│   ├── assets/
│   ├── components/
│   │   ├── atoms/
│   │   │   └── ...
│   │   ├── molecules/
│   │   │   └── ...
│   │   ├── organisms/
│   │   │   └── ...
│   │   ├── ThreeDWorld/
│   │   │   ├── Map.jsx
│   │   │   ├── Scene.jsx
│   │   │   └── ...
│   │   └── ...
│   ├── hooks/
│   ├── stories/
│   ├── utils/
│   ├── views/
│   │   ├── home/
│   │   │   ├── home.page.tsx
│   │   │   ├── home.template.tsx
│   │   │   └── home.css
│   │   └── ...
│   ├── App.jsx
│   ├── index.css
│   ├── main.jsx
├── .eslintignore
├── .eslintrc.cjs
├── .gitignore
├── .prettierrc
├── index.html
├── pnpm-lock.yaml
├── tsconfig.json
├── tsconfig.node.json
├── README.md
└── vite.config.ts
```
