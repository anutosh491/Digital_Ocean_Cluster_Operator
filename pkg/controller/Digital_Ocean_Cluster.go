package controller

import (
	"context"
	"log"
	"time"

	"github.com/anutosh491/Digital_Ocean_Cluster_Operator/pkg/apis/anutosh491.dev/v1alpha1"
	klientset "github.com/anutosh491/Digital_Ocean_Cluster_Operator/pkg/client/clientset/versioned"
	customscheme "github.com/anutosh491/Digital_Ocean_Cluster_Operator/pkg/client/clientset/versioned/scheme"
	kinf "github.com/anutosh491/Digital_Ocean_Cluster_Operator/pkg/client/informers/externalversions/anutosh491.dev/v1alpha1"
	klister "github.com/anutosh491/Digital_Ocean_Cluster_Operator/pkg/client/listers/anutosh491.dev/v1alpha1"
	"github.com/anutosh491/Digital_Ocean_Cluster_Operator/pkg/do"

	"github.com/kanisterio/kanister/pkg/poll"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	client kubernetes.Interface

	// clientset for custom resource Digital_Ocean_Cluster
	klient klientset.Interface
	// Digital_Ocean_Cluster has synced
	klusterSynced cache.InformerSynced
	// lister
	kLister klister.Digital_Ocean_ClusterLister
	// queue
	wq workqueue.RateLimitingInterface

	recorder record.EventRecorder
}

func NewController(client kubernetes.Interface, klient klientset.Interface, klusterInformer kinf.Digital_Ocean_ClusterInformer) *Controller {
	runtime.Must(customscheme.AddToScheme(scheme.Scheme))

	eveBroadCaster := record.NewBroadcaster()
	eveBroadCaster.StartStructuredLogging(0)
	eveBroadCaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{
		Interface: client.CoreV1().Events(""),
	})
	recorder := eveBroadCaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "Digital_Ocean_Cluster"})

	c := &Controller{
		client:        client,
		klient:        klient,
		klusterSynced: klusterInformer.Informer().HasSynced,
		kLister:       klusterInformer.Lister(),
		wq:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Digital_Ocean_Cluster"),
		recorder:      recorder,
	}

	klusterInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAdd,
			DeleteFunc: c.handleDel,
		},
	)

	return c
}

func (c *Controller) Run(ch chan struct{}) error {
	if ok := cache.WaitForCacheSync(ch, c.klusterSynced); !ok {
		log.Println("cache was not sycned")
	}

	go wait.Until(c.worker, time.Second, ch)

	<-ch
	return nil
}

func (c *Controller) worker() {
	for c.processNextItem() {

	}
}

func (c *Controller) processNextItem() bool {
	item, shutDown := c.wq.Get()
	if shutDown {
		// logs as well
		return false
	}

	defer c.wq.Forget(item)
	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		log.Printf("error %s calling Namespace key func on cache for item", err.Error())
		return false
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Printf("splitting key into namespace and name, error %s\n", err.Error())
		return false
	}

	Digital_Ocean_Cluster, err := c.kLister.Digital_Ocean_Clusters(ns).Get(name)
	if err != nil {
		log.Printf("error %s, Getting the Digital_Ocean_Cluster resource from lister", err.Error())
		return false
	}
	log.Printf("Digital_Ocean_Cluster spec that we have is %+v\n", Digital_Ocean_Cluster.Spec)

	clusterID, err := do.Create(c.client, Digital_Ocean_Cluster.Spec)
	if err != nil {
		// do something
		log.Printf("errro %s, creating the cluster", err.Error())
	}

	c.recorder.Event(Digital_Ocean_Cluster, corev1.EventTypeNormal, "ClusterCreation", "DO API was called to create the cluster")

	log.Printf("cluster id that we have is %s\n", clusterID)

	err = c.updateStatus(clusterID, "creating", Digital_Ocean_Cluster)
	if err != nil {
		log.Printf("error %s, updating status of the Digital_Ocean_Cluster %s\n", err.Error(), Digital_Ocean_Cluster.Name)
	}

	// query DO API to make sure clsuter' state is running
	err = c.waitForCluster(Digital_Ocean_Cluster.Spec, clusterID)
	if err != nil {
		log.Printf("error %s, waiting for cluster to be running", err.Error())
	}

	err = c.updateStatus(clusterID, "running", Digital_Ocean_Cluster)
	if err != nil {
		log.Printf("error %s updaring cluster status after waiting for cluster", err.Error())
	}

	c.recorder.Event(Digital_Ocean_Cluster, corev1.EventTypeNormal, "ClusterCreationCompleted", "DO Cluster creation was completed")
	return true
}

func (c *Controller) waitForCluster(spec v1alpha1.Digital_Ocean_ClusterSpec, clusterID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	return poll.Wait(ctx, func(ctx context.Context) (bool, error) {
		state, err := do.ClusterState(c.client, spec, clusterID)
		if err != nil {
			return false, err
		}
		if state == "running" {
			return true, nil
		}

		return false, nil
	})
}

func (c *Controller) updateStatus(id, progress string, Digital_Ocean_Cluster *v1alpha1.Digital_Ocean_Cluster) error {
	// get the latest version of Digital_Ocean_Cluster
	k, err := c.klient.anutosh491alpha1().Digital_Ocean_Clusters(Digital_Ocean_Cluster.Namespace).Get(context.Background(), Digital_Ocean_Cluster.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	k.Status.Digital_Ocean_ClusterID = id
	k.Status.Progress = progress
	_, err = c.klient.anutosh491alpha1().Digital_Ocean_Clusters(Digital_Ocean_Cluster.Namespace).UpdateStatus(context.Background(), k, metav1.UpdateOptions{})
	return err
}

func (c *Controller) handleAdd(obj interface{}) {
	log.Println("handleAdd was called")
	c.wq.Add(obj)
}

func (c *Controller) handleDel(obj interface{}) {
	log.Println("handleDel was called")
	c.wq.Add(obj)
}
