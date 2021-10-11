void buildRestd(String libc, String buildDir) {
    // sh "docker pull untangleinc/restd:build-${libc}"
    sh "docker-compose -f ${buildDir}/build/docker-compose.build.yml -p restd_${libc} run ${libc}-local"
    sh "cp ${buildDir}/cmd/restd/restd cmd/restd/restd-${libc}"
}

void archiveRestd() {
    archiveArtifacts artifacts:'cmd/restd/restd*', fingerprint: true
}

pipeline {
    agent none

    stages {
        stage('Build') {
            parallel {
                stage('Build musl') {
                    agent { label 'docker' }

                    environment {
                        libc = 'musl'
                        buildDir = "${env.HOME}/build-restd-${env.BRANCH_NAME}-${libc}/go/src/github.com/untangle/restd"
                    }

                    stages {
                        stage('Prep WS musl') {
                            steps { dir(buildDir) { checkout scm } }
                        }

                        stage('Build restd musl') {
                            steps {
                                buildRestd(libc, buildDir)
                                stash(name:"restd-${libc}", includes:"cmd/restd/restd*")
                            }
                        }
                    }

                    post {
                        success { archiveRestd() }
                    }
                }

                stage('Build glibc') {
                    agent { label 'docker' }

                    environment {
                        libc = 'glibc'
                        buildDir = "${env.HOME}/build-restd-${env.BRANCH_NAME}-${libc}/go/src/github.com/untangle/restd"
                    }

                    stages {
                        stage('Prep WS glibc') {
                            steps { dir(buildDir) { checkout scm } }
                        }

                        stage('Build restd glibc') {
                            steps {
                                buildRestd(libc, buildDir)
                                stash(name:"restd-${libc}", includes:'cmd/restd/restd*')
                            }
                        }
                    }

                    post {
                        success { archiveRestd() }
                    }
                }
            }
        }

        stage('Test') {
            parallel {
                stage('Test musl') {
                    agent { label 'docker' }

                    environment {
                        libc = 'musl'
                        restd = "cmd/restd/restd-${libc}"
                    }

                    stages {
                        stage('Prep musl') {
                            steps {
                                unstash(name:"restd-${libc}")
                            }
                        }

                        stage('File testing for musl') {
                            steps {
                                sh "test -f ${restd} && file ${restd} | grep -v -q GNU/Linux"
                            }
                        }
                    }
                }

                stage('Test libc') {
                    agent { label 'docker' }

                    environment {
                        libc = 'glibc'
                        restd = "cmd/restd/restd-${libc}"
                    }

                    stages {
                        stage('Prep libc') {
                            steps {
                                unstash(name:"restd-${libc}")
                            }
                        }

                        stage('File testing for libc') {
                            steps {
                                sh "test -f ${restd} && file ${restd} | grep -q GNU/Linux"
                            }
                        }
                        
                    }
                }
            }

            post {
                changed {
                    script {
                        // set result before pipeline ends, so emailer sees it
                        currentBuild.result = currentBuild.currentResult
                    }
                    emailext(to:'nfgw-engineering@untangle.com', subject:"${env.JOB_NAME} #${env.BUILD_NUMBER}: ${currentBuild.result}", body:"${env.BUILD_URL}")
                    slackSend(channel:'#team_engineering', message:"${env.JOB_NAME} #${env.BUILD_NUMBER}: ${currentBuild.result} at ${env.BUILD_URL}")
                }
            }
        }
    }
}
