pipeline {
    agent any

    parameters {
        choice(name: 'DEPLOY_ENV', choices: ['production', 'staging', 'development'], description: 'Target deployment environment')
        string(name: 'ROLLBACK_TAG', defaultValue: '', description: 'Build number of an earlier image to redeploy. Leave empty to build and deploy this commit.')
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
        FILE_DATA_BASE_PATH  = './filedata/0/'

        // Image tag: this build's number, or the build number given to roll back to
        IMAGE_NAME           = '2go-server'
        IMAGE_TAG            = "${params.ROLLBACK_TAG?.trim() ?: env.BUILD_NUMBER}"
        IMAGE_KEEP           = '5'

        // Backups of the database and file storage, kept in a Docker volume
        BACKUP_VOLUME        = '2go_backups'
        BACKUP_KEEP          = '10'

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
            when { expression { !params.ROLLBACK_TAG?.trim() } }
            steps {
                dir('server') {
                    sh 'go mod download'
                    sh 'go build -o main .'
                }
            }
        }

        stage('Test') {
            when { expression { !params.ROLLBACK_TAG?.trim() } }
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

        stage('Integration Test') {
            when { expression { !params.ROLLBACK_TAG?.trim() } }
            // Runs against a throwaway MySQL on its own Docker network, never the
            // production database. Tests run in a golang container on that network,
            // so this works whether Jenkins runs on the host or in a container.
            steps {
                sh '''#!/bin/bash
                    set -euo pipefail
                    NET="it-net-${BUILD_NUMBER}"
                    DB="it-mysql-${BUILD_NUMBER}"
                    RUNNER="it-go-${BUILD_NUMBER}"

                    docker network create "$NET"
                    docker run -d --name "$DB" --network "$NET" \
                        -e MYSQL_ROOT_PASSWORD=root \
                        -e MYSQL_DATABASE=testdb \
                        -e MYSQL_USER=testuser \
                        -e MYSQL_PASSWORD=password \
                        mysql:8.0

                    # Wait for the real server (the init server does not listen on TCP).
                    for i in $(seq 1 60); do
                        if docker exec "$DB" mysql -h127.0.0.1 -utestuser -ppassword -e 'SELECT 1' testdb >/dev/null 2>&1; then
                            break
                        fi
                        if [ "$i" -eq 60 ]; then
                            echo "Test MySQL did not start"; docker logs "$DB"; exit 1
                        fi
                        sleep 2
                    done

                    docker create --name "$RUNNER" --network "$NET" -w /server \
                        -e GO_ENV=test \
                        -e TEST_DB_DSN="testuser:password@tcp(${DB}:3306)/testdb?parseTime=true" \
                        -v 2go_gomodcache:/go/pkg/mod \
                        golang:1.25-bookworm \
                        go test -tags=integration ./internal/maintenance/test/integration/...

                    # server/ needs ../ocr for the go.mod replace directive.
                    tar --no-xattrs --exclude=server/main --exclude=server/.env --exclude=server/filedata -cf - server ocr | docker cp - "$RUNNER":/
                    docker start -a "$RUNNER"
                '''
            }
            post {
                always {
                    sh '''
                        docker rm -f "it-go-${BUILD_NUMBER}" "it-mysql-${BUILD_NUMBER}" >/dev/null 2>&1 || true
                        docker network rm "it-net-${BUILD_NUMBER}" >/dev/null 2>&1 || true
                    '''
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
FILE_DATA_BASE_PATH=${FILE_DATA_BASE_PATH}
ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
EOF
                    '''
                }
            }
        }

        stage('Docker Build') {
            when { expression { !params.ROLLBACK_TAG?.trim() } }
            steps {
                // Build context is the project root so Docker can reach both
                // server/ and ocr/ (required by the go.mod replace directive).
                sh 'docker build -f server/Dockerfile -t ${IMAGE_NAME}:${BUILD_NUMBER} -t ${IMAGE_NAME}:latest .'
            }
        }

        stage('Check Rollback Image') {
            when { expression { params.ROLLBACK_TAG?.trim() as boolean } }
            steps {
                sh '''
                    if ! docker image inspect "${IMAGE_NAME}:${IMAGE_TAG}" >/dev/null 2>&1; then
                        echo "Image ${IMAGE_NAME}:${IMAGE_TAG} not found. Available build numbers:"
                        docker images "${IMAGE_NAME}" --format '{{.Tag}}' | grep -E '^[0-9]+$' | sort -rn
                        exit 1
                    fi
                    echo "Rolling back to ${IMAGE_NAME}:${IMAGE_TAG}"
                '''
            }
        }

        stage('Backup') {
            // Dump the database and archive file storage before deploying, so a bad
            // release or a wrong repair can be restored. Fails the build (no deploy)
            // if a backup cannot be written.
            steps {
                sh '''#!/bin/bash
                    set -euo pipefail
                    TS=$(date +%Y%m%d-%H%M%S)

                    docker volume create "$BACKUP_VOLUME" >/dev/null

                    if docker ps --format '{{.Names}}' | grep -qx mysql_db; then
                        echo "Backing up database to db-$TS.sql.gz"
                        docker exec mysql_db sh -c 'MYSQL_PWD="$MYSQL_ROOT_PASSWORD" mysqldump -uroot --single-transaction --routines --no-tablespaces "$MYSQL_DATABASE"' \
                            | gzip \
                            | docker run -i --rm -v "$BACKUP_VOLUME":/backup alpine sh -c "cat > /backup/db-$TS.sql.gz && test -s /backup/db-$TS.sql.gz"
                    else
                        echo "mysql_db is not running (first deploy?); skipping database backup"
                    fi

                    if docker volume inspect server_filedata >/dev/null 2>&1; then
                        echo "Backing up file storage to filedata-$TS.tgz"
                        docker run --rm -v server_filedata:/data:ro -v "$BACKUP_VOLUME":/backup alpine \
                            tar czf "/backup/filedata-$TS.tgz" -C /data .
                    else
                        echo "server_filedata volume not found (first deploy?); skipping file backup"
                    fi

                    # Keep only the newest backups of each kind.
                    docker run --rm -v "$BACKUP_VOLUME":/backup alpine sh -c "
                        cd /backup
                        ls -1t db-*.sql.gz 2>/dev/null | tail -n +$((BACKUP_KEEP + 1)) | xargs -r rm -f
                        ls -1t filedata-*.tgz 2>/dev/null | tail -n +$((BACKUP_KEEP + 1)) | xargs -r rm -f
                        ls -lh
                    "
                '''
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

                        # Deploy IMAGE_TAG (read by docker-compose.prod.yml) and point latest at it,
                        # so a manual `docker-compose up` runs the same image.
                        docker tag "${IMAGE_NAME}:${IMAGE_TAG}" "${IMAGE_NAME}:latest"
                        docker-compose -p server -f docker-compose.prod.yml up -d --force-recreate --no-build api

                        # Keep the newest IMAGE_KEEP numbered images for rollback; never the deployed one.
                        docker images "${IMAGE_NAME}" --format '{{.Tag}}' | grep -E '^[0-9]+$' | sort -rn \
                            | tail -n +$((IMAGE_KEEP + 1)) | grep -vx "${IMAGE_TAG}" \
                            | xargs -r -I{} sh -c 'docker rmi "${IMAGE_NAME}:{}" || true'
                        docker image prune -f
                    '''
                }
            }
        }
    }

    post {
        success {
            echo "Build completed successfully! Deployed ${env.IMAGE_NAME}:${env.IMAGE_TAG}"
        }
        failure {
            echo 'Build failed!'
        }
    }
}
