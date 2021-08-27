# ModelMesh Proxy

ModelMesh Proxy leverages [gRPC-Gateway](https://github.com/grpc-ecosystem/grpc-gateway) to create a reverse-proxy server which translates a RESTful HTTP API into gRPC.

This allows sending inference requests to ModelMesh using REST.


## Prerequisites

- [ModelMesh Serving](https://github.com/kserve/modelmesh-serving) must be installed.

## Installation

```bash
kustomize build config | kubectl apply -n modelmesh-serving -f -
```

After installation you should see a new service and deployment:

```bash
kubectl get svc modelmesh-proxy  -n modelmesh-serving

NAME               TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
modelmesh-proxy    NodePort   10.101.137.44   <none>        8080:32189/TCP   2m

```

```bash
kubectl get deployment modelmesh-proxy

NAME              READY   UP-TO-DATE   AVAILABLE   AGE
modelmesh-proxy   1/1     1            1           2m
```

## Rest Inference

With the proxy installed, you can now perform external curl requests. For example, using the model and request data
from [here](https://github.com/kserve/modelmesh-serving/tree/main/docs#3-perform-a-grpc-inference-request), you can
now do something like:

```bash
curl -X POST -k http://192.168.49.2:32189/v2/models/example-mnist-predictor/infer -d '{ "inputs": [{ "name": "predict", "shape": [1, 64], "datatype": "FP32", "data": [0.0, 0.0, 1.0, 11.0, 14.0, 15.0, 3.0, 0.0, 0.0, 1.0, 13.0, 16.0, 12.0, 16.0, 8.0, 0.0, 0.0, 8.0, 16.0, 4.0, 6.0, 16.0, 5.0, 0.0, 0.0, 5.0, 15.0, 11.0, 13.0, 14.0, 0.0, 0.0, 0.0, 0.0, 2.0, 12.0, 16.0, 13.0, 0.0, 0.0, 0.0, 0.0, 0.0, 13.0, 16.0, 16.0, 6.0, 0.0, 0.0, 0.0, 0.0, 16.0, 16.0, 16.0, 7.0, 0.0, 0.0, 0.0, 0.0, 11.0, 13.0, 12.0, 1.0, 0.0]}]}'


{"model_name":"example-sklearn-mnist-svm__ksp-7702c1b55a","outputs":[{"name":"predict","datatype":"FP32","shape":[1],"data":[8]}]}
```

Note that the port is the NodePort of the `modelmesh-proxy` service, and the Node IP or Ingress Subdomain of your cluster.
