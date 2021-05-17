# Open Match POC
Simple match making using google open match.<br>
Currently it takes one ticket, creates a match (with only one player), assigns fake game server ip and sends the match

### Open match architecture<br>
![Image](https://open-match.dev/site/images/loam_create.png)

### Details
- Game Frontend here means our code backend
- Client means actual frontend

### How to run
1. Install [minikube](https://minikube.sigs.k8s.io/docs/start/)
2. Run minikube `minikube start --cpus=3 --memory=2500mb` (to delete use `minikube delete`)
3. Run `eval $(minikube docker-env)` (minikube will use local docker repository)
4. Cd into frontend and run `docker build -t realpvn/open-match-function .`<br>
(this builds match making function docker image)
5. Run setup `./setup --new`
    - It deletes old minikube and starts new one
    - It also sets required environment variable which will be used in our code to get host address and ports of our pods
6. Source bashrc file
7. Run frontend, and director. (Match function will run inside kubernetes as a service (refer to the image above) so we do not run it manually, it starts on step 5)
(Step 1-4 is required only first time after restart/installation of minikube, after that if we want to restart whole open-match pods/service use `./setup` and it should delete old pods/service and start new)
