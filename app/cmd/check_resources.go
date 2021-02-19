package cmd

import (
	"context"
	"fmt"
	"sync"

	"github.com/urfave/cli"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type CheckOpts struct {
	Client           *kubernetes.Clientset
	Logger           *zap.SugaredLogger
	ContextNamespace string
}

const checkResourcesDelta = 3

// NewScanCommand - create new instance of scan command.
func NewScanCommand(c CheckOpts) cli.Command {
	return cli.Command{
		Name:   "resources",
		Usage:  "Check resources readiness to production",
		Action: c.checkResources,
		Flags:  resourcesFlags(),
	}
}

func resourcesFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "namespace,n",
			Usage:  "Specify namespace name. Example [kubecheck scan -n default]",
			Hidden: true,
		},
		cli.BoolFlag{
			Name:  "ignore-security-context,i",
			Usage: "Ignore security context checks. Default: [true]",
		},
	}
}
func (o *CheckOpts) checkResources(ctx *cli.Context) {
	var ignore = ctx.Bool("ignore-security-context")
	namespace := o.getContextNamespace(ctx)
	var wg sync.WaitGroup

	wg.Add(checkResourcesDelta)

	go o.checkDeployments(namespace, ignore, &wg)
	go o.checkDaemonSets(namespace, ignore, &wg)
	go o.checkStatefulSets(namespace, ignore, &wg)

	wg.Wait()
}

func (o *CheckOpts) checkDeployments(namespace string, ignore bool, wg *sync.WaitGroup) {
	deployments, err := o.getNamespaceDeployments(namespace)
	if err != nil {
		o.Logger.Errorf("Failed to get deployments in namespace: %s", err)
		wg.Done()
	}
	for deployment := range deployments.Items {
		checkResources(deployments.Items[deployment].Name, "Deployment",
			ignore, &deployments.Items[deployment].Spec.Template.Spec)
	}
	wg.Done()
}

func (o *CheckOpts) checkDaemonSets(namespace string, ignore bool, wg *sync.WaitGroup) {
	daemonSets, err := o.getNamespaceDaemonSets(namespace)
	if err != nil {
		o.Logger.Errorf("Failed to get daemonSets in namespace: %s", err)
		wg.Done()
	}
	for daemonSet := range daemonSets.Items {
		checkResources(daemonSets.Items[daemonSet].Name, "DaemonSet", ignore,
			&daemonSets.Items[daemonSet].Spec.Template.Spec)
	}
	wg.Done()
}

func (o *CheckOpts) checkStatefulSets(namespace string, ignore bool, wg *sync.WaitGroup) {
	statefulSets, err := o.getNamespaceStatefulSets(namespace)
	if err != nil {
		o.Logger.Errorf("Failed to get deployments in namespace: %s", err)
		wg.Done()
	}
	for statefulSet := range statefulSets.Items {
		checkResources(statefulSets.Items[statefulSet].Name, "StatefulSet",
			ignore, &statefulSets.Items[statefulSet].Spec.Template.Spec)
	}
	wg.Done()
}

func checkResources(name, kind string, ignore bool, podSpec *v1.PodSpec) {
	for container := range podSpec.Containers {
		if podSpec.Containers[container].LivenessProbe == nil {
			fmt.Printf("[WARN] %s: %s, container: %s - does not have [LivenessProbe]\n", kind, name,
				podSpec.Containers[container].Name)
		}
		if podSpec.Containers[container].ReadinessProbe == nil {
			fmt.Printf("[WARN] %s: %s, container: %s - does not have [ReadinessProbe]\n", kind, name,
				podSpec.Containers[container].Name)
		}
		if podSpec.Containers[container].Resources.Limits == nil {
			fmt.Printf("[WARN] %s: %s, container: %s - does not have [Limits]\n", kind, name,
				podSpec.Containers[container].Name)
		}
		if podSpec.Containers[container].Resources.Requests == nil {
			fmt.Printf("[WARN] %s: %s, container: %s - does not have [Requests]\n", kind, name,
				podSpec.Containers[container].Name)
		}
		if podSpec.HostNetwork {
			fmt.Printf("[INFO] %s: %s, container: %s - has [Host Network]\n", kind, name,
				podSpec.Containers[container].Name)
		}
		if podSpec.HostPID {
			fmt.Printf("[WARN] %s: %s, container: %s - has [Host PID]\n", kind, name,
				podSpec.Containers[container].Name)
		}
		if podSpec.Containers[container].SecurityContext != nil && !ignore {
			checkContainerSecurityContext(name, kind, podSpec.Containers[container].Name,
				podSpec.Containers[container].SecurityContext)
		}
		if podSpec.Containers[container].SecurityContext == nil && !ignore {
			fmt.Printf("[ERROR] %s: %s, container: %s - has no provided [Security Context]\n", kind, name,
				podSpec.Containers[container].Name)
		}
	}
	if podSpec.SecurityContext != nil && !ignore {
		checkPodSecurityContext(name, kind, podSpec.SecurityContext)
	}
}

func checkContainerSecurityContext(name, kind, containerName string, containerSC *v1.SecurityContext) {
	if containerSC.AllowPrivilegeEscalation != nil {
		if *containerSC.AllowPrivilegeEscalation {
			fmt.Printf("[WARN] %s: %s, container: %s - has [Privilege Escalation]\n", kind, name, containerName)
		}
	}
	if containerSC.Privileged != nil {
		if *containerSC.Privileged {
			fmt.Printf("[CRIT] %s: %s, container: %s is [Privileged]\n", kind, name, containerName)
		}
	}
	if containerSC.RunAsGroup != nil {
		if *containerSC.RunAsGroup == 0 {
			fmt.Printf("[CRIT] %s: %s, container: %s - user has [Root Group]\n", kind, name, containerName)
		}
	}
	if containerSC.RunAsUser != nil {
		if *containerSC.RunAsUser == 0 {
			fmt.Printf("[CRIT] %s: %s, container: %s - user is [Root]\n", kind, name, containerName)
		}
	}
	if containerSC.ReadOnlyRootFilesystem != nil {
		if !*containerSC.ReadOnlyRootFilesystem {
			fmt.Printf("[WARN] %s: %s, container: %s - root filesystem read-write mounted\n", kind, name, containerName)
		}
	}
}

func checkPodSecurityContext(name, kind string, podSC *v1.PodSecurityContext) {
	if podSC.RunAsUser != nil {
		if *podSC.RunAsUser == 0 {
			fmt.Printf("[CRIT] %s: %s - user is [Root]\n", kind, name)
		}
	}
	if podSC.FSGroup != nil {
		if *podSC.FSGroup == 0 {
			fmt.Printf("[CRIT] %s: %s - FS group is [Root]\n", kind, name)
		}
	}
	if podSC.RunAsUser != nil {
		if *podSC.RunAsUser == 0 {
			fmt.Printf("[CRIT] %s: %s - user is [Root]\n", kind, name)
		}
	}
	if podSC.RunAsNonRoot != nil {
		if !*podSC.RunAsNonRoot {
			fmt.Printf("[CRIT] %s: %s - user run as [Root]\n", kind, name)
		}
	}
}

func (o *CheckOpts) getContextNamespace(ctx *cli.Context) string {
	var namespace = ctx.Parent().String("namespace")

	switch {
	case namespace != "":
		return namespace
	case ctx.String("namespace") != "":
		return ctx.String("namespace")
	case o.ContextNamespace != "":
		return o.ContextNamespace
	default:
		return "default"
	}
}

func (o *CheckOpts) getNamespaceDeployments(namespace string) (*appsv1.DeploymentList, error) {
	return o.Client.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
}

func (o *CheckOpts) getNamespaceStatefulSets(namespace string) (*appsv1.StatefulSetList, error) {
	return o.Client.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
}

func (o *CheckOpts) getNamespaceDaemonSets(namespace string) (*appsv1.DaemonSetList, error) {
	return o.Client.AppsV1().DaemonSets(namespace).List(context.TODO(), metav1.ListOptions{})
}
