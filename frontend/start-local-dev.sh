#!/bin/bash

: <<'END_COMMENT'
1. make sure the maxmind files are in the maxmind-db directory in the root folder
2. grant permission to the script by running chmod +x start-local-dev.sh
3. run the script by running ./start-local-dev.sh
END_COMMENT



cd .. || { echo "Failed to change directory to root"; exit 1; }


# Run the Aztec P2P Explorer server in the background with nohup
nohup ./aztec-p2p-explorer server --maxmind-dir maxmind-db >  /dev/null 2>&1 & # for log file do: > backend.log 2>&1 &
# Check if the command was successful
if [ $? -eq 0 ]; then
    echo "Aztec P2P Explorer Backend server started successfully."
else
    echo "Failed to start Aztec P2P Explorer Backend server."
    exit 1
fi

cd frontend/ || { echo "Failed to change directory to frontend/"; exit 1; }

pnpm run dev --host

if [ $? -ne 0 ]; then
    echo "Failed to start frontend development server."
    exit 1
fi
