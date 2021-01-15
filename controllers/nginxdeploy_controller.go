/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s/util"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/runtime"
	batchv1 "k8s/api/v1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// NginxDeployReconciler reconciles a NginxDeploy object
type NginxDeployReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	NginxDeployFinalizer = "nginxdeploy.finalizers.batch.my.domain"
)


// +kubebuilder:rbac:groups=batch.my.domain,resources=nginxdeploys,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.my.domain,resources=nginxdeploys/status,verbs=get;update;patch
// 当控制器检测到自定义CR发生变化时调用该函数
func (r *NginxDeployReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("nginxdeploy", req.NamespacedName)

	// your logic here
	log.Info(req.NamespacedName.Namespace)
	obj := &batchv1.NginxDeploy{}
	// 查找自定义CR对象
	err := r.Get(ctx,req.NamespacedName,obj)
	if err != nil{
		if err := client.IgnoreNotFound(err); err == nil {
			// 没有找到对应的CR说明已经删除了对应的CR,此处进入被删除成功后的生命周期
			log.Info("没有找到对应的CR说明此次变化事件应该是删除了对应的CR进入被删除成功后的生命周期")
			return ctrl.Result{}, nil
		} else {
			log.Error(err, "不是未找到的错误，那么就是意料之外的错误，所以这里直接返回错误")
			return ctrl.Result{}, err
		}
	}
	// 到此处说明已接收到资源更新变化，为确保删除CR时能够级联删除CR管理的所有子对象，需要添加如下判断
	//前台级联删除（Foreground Cascading Deletion）:
	//在这种删除策略中，所有者对象的删除将会持续到其所有从属对象都被删除为止。当所有者被删除时，会进入“正在删除”（deletion in progress）状态，此时：
	//对象仍然可以通过 REST API 查询到（可通过 kubectl 或 kuboard 查询到）
	//对象的 deletionTimestamp 字段被设置
	//对象的 metadata.finalizers 包含值 foregroundDeletion
	//一旦对象被设置为 “正在删除” 状态，垃圾回收器将删除其从属对象。当垃圾回收器已经删除了所有的“blocking”从属对象（ownerReference.blockOwnerDeletion=true 的对象）以后，将删除所有者对象。
	// 回收机制官网介绍：https://kubernetes.io/zh/docs/concepts/workloads/controllers/garbage-collection/
	if obj.ObjectMeta.DeletionTimestamp.IsZero(){
		// 如果对象的ObjectMeta.DeletionTimestamp字段设置为空，说明此监听到的变化是新增而不是删除，此时应该创建对应的pod
		if !util.ContainsString(obj.ObjectMeta.Finalizers, NginxDeployFinalizer) {
			obj.ObjectMeta.Finalizers = append(obj.ObjectMeta.Finalizers, NginxDeployFinalizer)
			if err := r.Update(ctx, obj); err != nil {
				return ctrl.Result{}, err
			}
		}
		if _, err := r.applyDeployment(ctx, req, obj); err != nil {
			return ctrl.Result{}, nil
		}
	}else {
		log.Info("进入到删除这个 DemoMicroService CR 的逻辑")
		if util.ContainsString(obj.ObjectMeta.Finalizers, NginxDeployFinalizer) {
			log.Info("如果 finalizers 被清空，则该 DemoMicroService CR 就已经不存在了，所以必须在次之前删除 Deployment")
			if err := r.cleanDeployment(ctx, req); err != nil {
				return ctrl.Result{}, nil
			}
		}
		log.Info("清空 finalizers，在此之后该 DemoMicroService CR 才会真正消失")
		obj.ObjectMeta.Finalizers = util.RemoveString(obj.ObjectMeta.Finalizers, NginxDeployFinalizer)
		if err := r.Update(ctx, obj); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *NginxDeployReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.NginxDeploy{}).
		Complete(r)
}


func (r *NginxDeployReconciler) applyDeployment(ctx context.Context, req ctrl.Request, obj *batchv1.NginxDeploy) (*appv1.Deployment, error) {
	podLabels := map[string]string{
		"app": req.Name,
	}
	deployment := appv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Spec: appv1.DeploymentSpec{
			Replicas: &obj.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            req.Name,
							Image:           obj.Spec.Image,
							ImagePullPolicy: "Always",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	oldDeployment := &appv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, oldDeployment); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// 如果 deployment 不存在，则创建它
			if err := r.Create(ctx, &deployment); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	// 此时表示 deployment 已经存在，则更新它
	if err := r.Update(ctx, &deployment); err != nil {
		return nil, err
	}
	return &deployment, nil
}

func (r *NginxDeployReconciler) cleanDeployment(ctx context.Context, req ctrl.Request) error {
	deployment := &appv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, deployment); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// 既然已经没了，do nothing
			return nil
		} else {
			return err
		}
	}
	if err := r.Delete(ctx, deployment); err != nil {
		return err
	}
	return nil
}
