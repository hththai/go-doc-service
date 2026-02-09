pipeline {
    agent any

    tools {
        go '1.25.0'
    }

    environment {
        API_DOMAIN_LOCAL = 'api.golang.localdomain'
        API_DOMAIN_PROD  = 'api.golang.hthai.cloud'
        API_PORT         = '8088'
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

        stage('Docker Build') {
            steps {
                dir('server') {
                    sh 'docker build -t 2go-server:latest .'
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
