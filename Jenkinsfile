pipeline {
    agent any

    tools {
        go '1.25.0'
    }

    environment {
        // Non-sensitive config
        API_DOMAIN_PROD  = 'box.hthai.cloud'
        API_PORT         = '8088'
        GO_ENV           = 'production'

        // Secrets from Jenkins Credentials
        DB_ROOT_PASSWORD = credentials('db-root-password')
        DB_USER          = credentials('db-user')
        DB_PASSWORD      = credentials('db-password')
        JWT_SECRET       = credentials('jwt-secret')
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
            steps {
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
EOF
                    '''
                }
            }
        }

        stage('Docker Build') {
            steps {
                dir('server') {
                    sh 'docker build -t 2go-server:latest .'
                }
            }
        }

        stage('Deploy') {
            steps {
                dir('server') {
                    sh '''
                        export API_DOMAIN=${API_DOMAIN_PROD}
                        export DB_NAME=goDocument
                        docker-compose -f docker-compose.prod.yml down --remove-orphans || true
                        docker-compose -f docker-compose.prod.yml up -d --build
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
