make build
export JWT_SECRET=$(openssl rand -hex 32)
./bin/taskflow