# Description
Kubecheck will check all your pods for basic k8s requirements such as SecurityContexts/Probes/Requests/etc.

# Install
```
make build
```

#### Check Resources 
```
kubecheck resources -n test
   [WARN] Pod: proxy-c4ccb7fdc-9tdjc, container: app does not has LivenessProbe
   [WARN] Pod: proxy-c4ccb7fdc-9tdjc, container: app does not has Limits
   [WARN] Pod: proxy-c4ccb7fdc-9tdjc, container: app does not has Requests
```

### Environment Variables
|  Env  |  Value |
| ------|------- |
| KUBECONFIG_PATH | /home/$USER/.kube/config |
