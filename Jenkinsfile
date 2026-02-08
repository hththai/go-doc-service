pipeline {
    agent any

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
