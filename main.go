package main

import (
	"flag"
	clientset "k8s-custom-crd-controller/pkg/generated/clientset/versioned"
	informers "k8s-custom-crd-controller/pkg/generated/informers/externalversions"
	"k8s-custom-crd-controller/pkg/signals"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"time"

	"k8s.io/klog/v2"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	klog.InitFlags(nil)
	flag.Parsed()

	ctx := signals.SetupSignalHandler()
	logger := klog.FromContext(ctx)

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		logger.Error(err, "Error building kubeconfig")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		logger.Error(err, "Error building kubernetes clientset")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	networkClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		logger.Error(err, "Error building example clientset")
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	networkInformerFactory := informers.NewSharedInformerFactory(networkClient, time.Second*30)

	controller := NewController(ctx, kubeClient, networkClient, kubeInformerFactory.Apps().V1().Deployments(), networkInformerFactory.Samplecrd().V1().Networks())

	kubeInformerFactory.Start(ctx.Done())
	networkInformerFactory.Start(ctx.Done())
	if err = controller.Run(ctx, 2); err != nil {
		logger.Error(err, "Error running controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server (overrides any value in kubeconfig. Only required if out-of-cluster.)")
}
