pipeline {
    agent any

    parameters {
        choice(name: 'DEPLOY_ENV', choices: ['production', 'staging', 'development'], description: 'Target deployment environment')
    }

    tools {
        go '1.25.0'
    }

    environment {
        // Derived from parameter
        GO_ENV           = "${params.DEPLOY_ENV}"

        // Non-sensitive config
        API_DOMAIN_PROD      = 'box.hthai.cloud'
        API_PORT             = '8088'
        COOKIE_SECURE        = "${params.DEPLOY_ENV != 'development' ? 'true' : 'false'}"
        RATE_LIMIT_REQUESTS  = '100'

        // Secrets from Jenkins Credentials
        DB_ROOT_PASSWORD   = credentials('db-root-password')
        DB_USER            = credentials('db-user')
        DB_PASSWORD        = credentials('db-password')
        JWT_SECRET         = credentials('jwt-secret')
        ANTHROPIC_API_KEY  = credentials('anthropic-api-key')
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Build') {
            steps {
                dir('server') {
                    sh 'go mod download'
                    sh 'go build -o main .'
                }
            }
        }

        stage('Test') {
            environment {
                GO_ENV = 'test'
            }
            steps {
                // go-fitz (used by the ocr package) loads libmupdf.so at init
                // time via purego. Install it so the shared library is present
                // on the agent even though the tests themselves use a mock.
                sh 'apt-get install -y --no-install-recommends libmupdf-dev'
                dir('server') {
                    sh 'go test -v ./...'
                }
            }
        }

        stage('Generate Config') {
            steps {
                dir('server/dynamic') {
                    sh 'envsubst < config.yml.template > config.yml'
                    sh 'cat config.yml'
                }
            }
        }

        stage('Generate .env') {
            steps {
                dir('server') {
                    sh '''
                        cat > .env << EOF
DB_ROOT_PASSWORD=${DB_ROOT_PASSWORD}
DB_NAME=goDocument
DB_USER=${DB_USER}
DB_PASSWORD=${DB_PASSWORD}
DB_HOST=db
DB_PORT=3306
GO_ENV=${GO_ENV}
API_PORT=${API_PORT}
JWT_SECRET=${JWT_SECRET}
CORS_ORIGINS=https://${API_DOMAIN_PROD}
COOKIE_SECURE=${COOKIE_SECURE}
RATE_LIMIT_REQUESTS=${RATE_LIMIT_REQUESTS}
ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
EOF
                    '''
                }
            }
        }

        stage('Docker Build') {
            steps {
                // Build context is the project root so Docker can reach both
                // server/ and ocr/ (required by the go.mod replace directive).
                sh 'docker build -f server/Dockerfile -t 2go-server:latest .'
            }
        }

        stage('Deploy') {
            steps {
                dir('server') {
                    sh '''
                        export API_DOMAIN=${API_DOMAIN_PROD}
                        export DB_NAME=goDocument

                        # One-time migration: move data from old mydata volume to new named volumes
                        # Docker Compose prefixes volumes with the project (directory) name: server_mydata
                        if docker volume inspect server_mydata >/dev/null 2>&1; then
                            echo "Migrating data from old server_mydata volume..."
                            docker run --rm \
                                -v server_mydata:/old \
                                -v server_filedata:/new_filedata \
                                -v server_logdata:/new_logdata \
                                alpine sh -c "
                                    cp -rp /old/filedata/. /new_filedata/ 2>/dev/null || true
                                    mkdir -p /new_logdata
                                    cp -rp /old/app/log/. /new_logdata/ 2>/dev/null || true
                                "
                            echo "Migration done. Remove old volume manually when ready: docker volume rm server_mydata"
                        fi

                        docker-compose -f docker-compose.prod.yml up -d --force-recreate --no-build api
                        docker image prune -f
                    '''
                }
            }
        }
    }

    post {
        success {
            echo 'Build completed successfully!'
        }
        failure {
            echo 'Build failed!'
        }
    }
}
