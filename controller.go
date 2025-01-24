package main

import (
	"context"
	"fmt"
	"golang.org/x/time/rate"
	samplecrdv1 "k8s-custom-crd-controller/pkg/apis/samplecrd/v1"
	clientset "k8s-custom-crd-controller/pkg/generated/clientset/versioned"
	networkscheme "k8s-custom-crd-controller/pkg/generated/clientset/versioned/scheme"
	informers "k8s-custom-crd-controller/pkg/generated/informers/externalversions/samplecrd/v1"
	listers "k8s-custom-crd-controller/pkg/generated/listers/samplecrd/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"time"
)

const controllerAgentName = "network-controller"

const (
	SuccessSynced         = "Synced"
	ErrResourceExists     = "ErrResourceExists"
	MessageResourceExists = "Resource %q already exists and is not managed by Network"
	MessageResourceSynced = "Network synced successfully"
	FieldManager          = controllerAgentName
)

type Controller struct {
	kubeclientset    kubernetes.Interface
	networkclientset clientset.Interface

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	networkLister     listers.NetworkLister
	networkSynced     cache.InformerSynced

	workqueue workqueue.TypedRateLimitingInterface[cache.ObjectName]
	recorder  record.EventRecorder
}

func NewController(
	ctx context.Context,
	kubeclientset kubernetes.Interface,
	networkclientset clientset.Interface,
	deploymentInformer appsinformers.DeploymentInformer,
	networkInformer informers.NetworkInformer) *Controller {
	logger := klog.FromContext(ctx)

	utilruntime.Must(networkscheme.AddToScheme(scheme.Scheme))
	logger.V(4).Info("Creating event broadcaster")

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	ratelimiter := workqueue.NewTypedMaxOfRateLimiter(
		workqueue.NewTypedItemExponentialFailureRateLimiter[cache.ObjectName](5*time.Microsecond, 1000*time.Second),
		&workqueue.TypedBucketRateLimiter[cache.ObjectName]{rate.NewLimiter(rate.Limit(50), 300)})

	controller := &Controller{
		kubeclientset:     kubeclientset,
		networkclientset:  networkclientset,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		networkLister:     networkInformer.Lister(),
		networkSynced:     networkInformer.Informer().HasSynced,
		workqueue:         workqueue.NewTypedRateLimitingQueue(ratelimiter),
		recorder:          recorder,
	}
	logger.Info("Setting up event handlers")
	networkInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueNetwork,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueNetwork(new)
		},
	})
	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*appsv1.Deployment)
			oldDepl := old.(*appsv1.Deployment)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})
	return controller
}

func (c *Controller) Run(ctx context.Context, workers int) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()
	logger := klog.FromContext(ctx)

	logger.Info("Starting Network controller")
	logger.Info("Waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(ctx.Done(), c.deploymentsSynced, c.networkSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	logger.Info("Starting workers")
	<-ctx.Done()
	logger.Info("Stopping workers")

	return nil
}

func (c *Controller) processNextWorkItem(ctx context.Context) bool {
	objRef, shutdown := c.workqueue.Get()
	logger := klog.FromContext(ctx)

	if shutdown {
		return false
	}

	defer c.workqueue.Done(objRef)

	err := c.syncHandler(ctx, objRef)
	if err == nil {
		c.workqueue.Forget(objRef)
		logger.Info("Successfully synced", "objectName", objRef)
		return true
	}

	utilruntime.HandleErrorWithContext(ctx, err, "Error syncing; requeuing for later retry", "objectReference", objRef)
	c.workqueue.AddRateLimited(objRef)
	return true
}

func (c *Controller) syncHandler(ctx context.Context, objectRef cache.ObjectName) error {
	logger := klog.FromContext(ctx)

	network, err := c.networkLister.Networks(objectRef.Namespace).Get(objectRef.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleErrorWithContext(ctx, err, "Network referenced by item in work queue no longer exists", "objectReference", objectRef)
			return nil
		}
		return err
	}
	deploymentName := network.Spec.DeploymentName
	if deploymentName == "" {
		utilruntime.HandleErrorWithContext(ctx, nil, "Deployment name must be specified", "objectReference", objectRef)
		return nil
	}
	deployment, err := c.deploymentsLister.Deployments(network.Namespace).Get(deploymentName)
	if errors.IsNotFound(err) {
		deployment, err = c.kubeclientset.AppsV1().Deployments(network.Namespace).Create(ctx, newDeployment(network), metav1.CreateOptions{FieldManager: FieldManager})
	}
	if err != nil {
		return err
	}

	if !metav1.IsControlledBy(deployment, network) {
		msg := fmt.Sprintf(MessageResourceExists, deployment.Name)
		c.recorder.Event(network, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf("%s", msg)
	}

	if network.Spec.Replicas != nil && *network.Spec.Replicas != *deployment.Spec.Replicas {
		logger.V(4).Info("Update deployment replicas", "currentReplicas", *deployment.Spec.Replicas, "desiredReplicas", *network.Spec.Replicas)
		deployment, err = c.kubeclientset.AppsV1().Deployments(network.Namespace).Update(ctx, newDeployment(network), metav1.UpdateOptions{FieldManager: FieldManager})
	}

	if err != nil {
		return err
	}

	err = c.updateNetworkStatus(ctx, network, deployment)
	if err != nil {
		return err
	}

	c.recorder.Event(network, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateNetworkStatus(ctx context.Context, network *samplecrdv1.Network, deployment *appsv1.Deployment) error {
	networkCopy := network.DeepCopy()
	networkCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	_, err := c.networkclientset.SamplecrdV1().Networks(network.Namespace).UpdateStatus(ctx, networkCopy, metav1.UpdateOptions{FieldManager: FieldManager})

	return err
}

func newDeployment(network *samplecrdv1.Network) *appsv1.Deployment {
	labels := map[string]string{
		"app":        "network-controller",
		"controller": network.Name,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      network.Name,
			Namespace: network.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(network, samplecrdv1.SchemeGroupVersion.WithKind("Network")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: network.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx-network-controller",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}
}

func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	logger := klog.FromContext(context.Background())
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleErrorWithContext(context.Background(), nil, "error decoding object, invalid type", "type", fmt.Sprintf("%T", obj))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleErrorWithContext(context.Background(), nil, "error decoding object tombstone, invalid type", "type", fmt.Sprintf("%T", tombstone.Obj))
			return
		}
		logger.V(4).Info("Recovered deleted object", "resourceName", object.GetName())
	}
	logger.V(4).Info("Processing object", "object", klog.KObj(object))
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		if ownerRef.Kind != "Network" {
			return
		}
		network, err := c.networkLister.Networks(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			logger.V(4).Info("Ignore orphaned object", "object", klog.KObj(object), "network", ownerRef.Name)
			return
		}
		c.enqueueNetwork(network)
		return
	}
}

func (c *Controller) enqueueNetwork(obj interface{}) {
	if objectRef, err := cache.ObjectToName(obj); err == nil {
		utilruntime.HandleError(err)
		return
	} else {
		c.workqueue.Add(objectRef)
	}
}
