make build
export JWT_SECRET=$(openssl rand -hex 32)
export API_BASE_PATH='/taskflow/api'
./bin/taskflow start
