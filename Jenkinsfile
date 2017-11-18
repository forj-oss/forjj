pipeline {
    agent any

    stages {
        stage('Build') {
            steps {
                withEnv(["DOCKER_JENKINS_HOME=${env.DOCKER_JENKINS_MOUNT}"]) {
                    sh('''source ./build-env.sh
                    build.sh''')
                }
            }
        }
        stage('Tests') {
            steps {
                sh('''source ./build-env.sh
                go test''')
            }
        }
        stage('Deploy') {
            when { branch 'master' }
            steps {
                withCredentials([string(credentialsId: 'github-token', variable: 'GITHUB_TOKEN')]) {
                    sh('''source ./build-env.sh
                    publish.sh latest''')
                }
            }
        }
    }
}
