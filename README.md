# Digital_Ocean_Cluster_Operator
A Kubernetes operator to deploy a Cluster on Digital Ocean dynamically.


### Use Case
Let's say we want to provision a Kubernetes cluster on Digital Ocean .... For that we would need to look for the required fields for creating a clusteer (That can be done through this link https://docs.digitalocean.com/reference/api/api-reference/#operation/kubernetes_create_cluster) .So we can see that we essentially need 4 fields to create a cluster. Those are name, version, region and node pools.
Now obviously we know that we don't have any k8s native resource where we can specify these fields , hence we would have to create a custom resource with the help of CRDs . Once we have our CR and CRD in place, we can create an operator that would listen on deletion/addition of those CRs and do certain operations based on that.


### Basic Building Blocks for the operator
1) We would have to let the Kube-API server know that we want to support a new type through a CRD. 
2) We also know that every k8s object has a Group and a Version.So let's name our custom resource type as `Digitial_Ocean_Cluster` and our Group as `anutosh491.dev` and Version would be `v1alpha1`.
3) Now we would want to register our type/resource as a K8s resource using the `runtime`, `runtime/schema` and the `apis/meta/v1` package . The functions used are `addKnownTypes` from runtime package and `AddToGroupVersion` from metav1.
4) We would also have to generate some code .This comes in handy for us to be able to register our resource as a K8S resource . For doing that our resource should implement a deep copy function, that we currently won't have. Other than that, we also need code for some features that we currently lack, for eg
  - We lack a client set for our custom resource.
  - We also lack informers and listers for our custom resource.
For this we would be using the code generator project.
5) Lastly , now that we've successfully registered our resource we should make a CRD for our resource.

