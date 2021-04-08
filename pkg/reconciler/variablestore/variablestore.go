/*
Copyright 2019 The Knative Authors

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

package variablestore

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	runreconciler "github.com/tektoncd/pipeline/pkg/client/injection/reconciler/pipeline/v1alpha1/run"
	listersalpha "github.com/tektoncd/pipeline/pkg/client/listers/pipeline/v1alpha1"
	variablestorev1alpha1 "github.com/vincentpli/cel-tekton/pkg/apis/variablestores/v1alpha1"
	variableclientset "github.com/vincentpli/cel-tekton/pkg/client/clientset/versioned"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
	"knative.dev/pkg/tracker"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"

	"github.com/tektoncd/pipeline/pkg/reconciler/events"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Reconciler implements addressableservicereconciler.Interface for
// AddressableService resources.
type Reconciler struct {
	// Tracker builds an index of what resources are watching other resources
	// so that we can immediately react to changes tracked resources.
	Tracker tracker.Interface

	//Clientset about resources
	variablestoreClientSet variableclientset.Interface

	// Listers index properties about resources
	runLister listersalpha.RunLister
}

// Check that our Reconciler implements Interface
var _ runreconciler.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, run *v1alpha1.Run) reconciler.Event {
	var merr error
	logger := logging.FromContext(ctx)
	logger.Infof("Reconciling Run %s/%s at %v", run.Namespace, run.Name, time.Now())

	// Check that the Run references a Exception CRD.  The logic is controller.go should ensure that only this type of Run
	// is reconciled this controller but it never hurts to do some bullet-proofing.
	if run.Spec.Ref == nil ||
		run.Spec.Ref.APIVersion != variablestorev1alpha1.SchemeGroupVersion.String() ||
		run.Spec.Ref.Kind != "VariableStore" {
		logger.Errorf("Received control for a Run %s/%s that does not reference a VariableStore custom CRD", run.Namespace, run.Name)
		return nil
	}

	// If the Run has not started, initialize the Condition and set the start time.
	if !run.HasStarted() {
		logger.Infof("Starting new Run %s/%s", run.Namespace, run.Name)
		run.Status.InitializeConditions()
		// In case node time was not synchronized, when controller has been scheduled to other nodes.
		if run.Status.StartTime.Sub(run.CreationTimestamp.Time) < 0 {
			logger.Warnf("Run %s createTimestamp %s is after the Run started %s", run.Name, run.CreationTimestamp, run.Status.StartTime)
			run.Status.StartTime = &run.CreationTimestamp
		}

		// Emit events. During the first reconcile the status of the Run may change twice
		// from not Started to Started and then to Running, so we need to sent the event here
		// and at the end of 'Reconcile' again.
		// We also want to send the "Started" event as soon as possible for anyone who may be waiting
		// on the event to perform user facing initialisations, such has reset a CI check status
		afterCondition := run.Status.GetCondition(apis.ConditionSucceeded)
		events.Emit(ctx, nil, afterCondition, run)
	}

	if run.IsDone() {
		logger.Infof("Run %s/%s is done", run.Namespace, run.Name)
		return nil
	}

	// Store the condition before reconcile
	beforeCondition := run.Status.GetCondition(apis.ConditionSucceeded)

	// Reconcile the Run
	if err := r.reconcile(ctx, run); err != nil {
		logger.Errorf("Reconcile error: %v", err.Error())
		merr = multierror.Append(merr, err)
	}

	afterCondition := run.Status.GetCondition(apis.ConditionSucceeded)
	events.Emit(ctx, beforeCondition, afterCondition, run)

	// Only transient errors that should retry the reconcile are returned.
	return merr
}

func (r *Reconciler) reconcile(ctx context.Context, run *v1alpha1.Run) error {
	logger := logging.FromContext(ctx)
	variablestore, err := r.getVariableStore(ctx, run)
	if err != nil {
		logger.Errorf("Error retrieving VariableStore for Run %s/%s: %s", run.Namespace, run.Name, err)
		run.Status.MarkRunFailed(variablestorev1alpha1.VariableStoreReasonCouldntGet.String(),
			"Error retrieving VariableStore for Run %s/%s: %s",
			run.Namespace, run.Name, err)
		return nil
	}

	if err := validate(run); err != nil {
		logger.Errorf("Run %s/%s is invalid because of %s", run.Namespace, run.Name, err)
		run.Status.MarkRunFailed(variablestorev1alpha1.ReasonFailedValidation.String(),
			"Run can't be run because it has an invalid spec - %v", err)
		return nil
	}

	// Create a program environment configured with the standard library of CEL functions and macros
	env, err := cel.NewEnv(cel.Declarations())
	if err != nil {
		logger.Errorf("Couldn't create a program env with standard library of CEL functions & macros when reconciling Run %s/%s: %v", run.Namespace, run.Name, err)
		return err
	}

	var runResults []v1alpha1.RunResult
	contextExpressions := map[string]interface{}{}

	// If refrenced VariableStore not null, all variables in that will be the context
	if variablestore != nil {
		for _, variable := range variablestore.Spec.Vars {
			contain, _ := containsVar(variable.Name, run.Spec.Params)
			if contain {
				continue
			}

			contextExpressions[variable.Name] = variable.Value
			env, err = env.Extend(cel.Declarations(decls.NewVar(variable.Name, decls.Any)))
			if err != nil {
				logger.Errorf("CEL expression %s could not be add to context env when reconciling Run %s/%s: %v", variable.Name, run.Namespace, run.Name, err)
				run.Status.MarkRunFailed(variablestorev1alpha1.ReasonEvaluationError.String(),
					"CEL expression %s could not be add to context env", variable.Name, err)
				return nil
			}
		}
	}

	for _, param := range run.Spec.Params {
		// Combine the Parse and Check phases CEL program compilation to produce an Ast and associated issues
		ast, iss := env.Compile(param.Value.StringVal)
		if iss.Err() != nil {
			logger.Errorf("CEL expression %s could not be parsed when reconciling Run %s/%s: %v", param.Name, run.Namespace, run.Name, iss.Err())
			run.Status.MarkRunFailed(variablestorev1alpha1.ReasonSyntaxError.String(),
				"CEL expression %s could not be parsed", param.Name, iss.Err())
			return nil
		}

		// Generate an evaluable instance of the Ast within the environment
		prg, err := env.Program(ast)
		if err != nil {
			logger.Errorf("CEL expression %s could not be evaluated when reconciling Run %s/%s: %v", param.Name, run.Namespace, run.Name, err)
			run.Status.MarkRunFailed(variablestorev1alpha1.ReasonEvaluationError.String(),
				"CEL expression %s could not be evaluated", param.Name, err)
			return nil
		}

		// Evaluate the CEL expression (Ast)
		out, _, err := prg.Eval(contextExpressions)
		if err != nil {
			logger.Errorf("CEL expression %s could not be evaluated when reconciling Run %s/%s: %v", param.Name, run.Namespace, run.Name, err)
			run.Status.MarkRunFailed(variablestorev1alpha1.ReasonEvaluationError.String(),
				"CEL expression %s could not be evaluated", param.Name, err)
			return nil
		}

		// Evaluation of CEL expression was successful
		logger.Infof("CEL expression %s evaluated successfully when reconciling Run %s/%s", param.Name, run.Namespace, run.Name)
		runResults = append(runResults, v1alpha1.RunResult{
			Name:  param.Name,
			Value: fmt.Sprintf("%s", out.ConvertToType(types.StringType).Value()),
		})
		contextExpressions[param.Name] = fmt.Sprintf("%s", out.ConvertToType(types.StringType).Value())
		env, err = env.Extend(cel.Declarations(decls.NewVar(param.Name, decls.Any)))
		if err != nil {
			logger.Errorf("CEL expression %s could not be add to context env when reconciling Run %s/%s: %v", param.Name, run.Namespace, run.Name, err)
			run.Status.MarkRunFailed(variablestorev1alpha1.ReasonEvaluationError.String(),
				"CEL expression %s could not be add to context env", param.Name, err)
			return nil
		}

		//Append calculated variables to VariableStore
		if variablestore != nil {
			variable := variablestorev1alpha1.Var{
				Name:  param.Name,
				Value: fmt.Sprintf("%s", out.ConvertToType(types.StringType).Value()),
			}

			contain, index := containsParam(param.Name, variablestore.Spec.Vars)
			if contain {
				variablestore.Spec.Vars[index] = variable
			} else {
				variablestore.Spec.Vars = append(variablestore.Spec.Vars, variable)
			}
		}

	}

	// Update VariableStore ??? variablestore do not existed
	if variablestore != nil {
		_, err = r.variablestoreClientSet.CustomV1alpha1().VariableStores(run.Namespace).Update(ctx, variablestore, metav1.UpdateOptions{})
		if err != nil {
			logger.Errorf("Update VariableStore: %s hit excetion: %v", variablestore.Name, err)
			run.Status.MarkRunFailed(variablestorev1alpha1.VariableStoreReasonUpdateFaild.String(),
				"Update VariableStore: %s hit excetion: %v", variablestore.Name, err)
			return nil
		}
	}

	run.Status.Results = append(run.Status.Results, runResults...)
	run.Status.MarkRunSucceeded(variablestorev1alpha1.ReasonEvaluationSuccess.String(),
		"CEL expressions were evaluated successfully")

	return nil
}

func (r *Reconciler) getVariableStore(ctx context.Context, run *v1alpha1.Run) (*variablestorev1alpha1.VariableStore, error) {
	var variablestore *variablestorev1alpha1.VariableStore

	if run.Spec.Ref != nil && run.Spec.Ref.Name != "" {
		// Use the k8 client to get the TaskLoop rather than the lister.  This avoids a timing issue where
		// the TaskLoop is not yet in the lister cache if it is created at nearly the same time as the Run.
		// See https://github.com/tektoncd/pipeline/issues/2740 for discussion on this issue.
		//
		vs, err := r.variablestoreClientSet.CustomV1alpha1().VariableStores(run.Namespace).Get(ctx, run.Spec.Ref.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		variablestore = vs
	}

	return variablestore, nil
}

func validate(run *v1alpha1.Run) (errs *apis.FieldError) {
	errs = errs.Also(validateExpressionsProvided(run))
	errs = errs.Also(validateExpressionsType(run))
	return errs
}

func validateExpressionsProvided(run *v1alpha1.Run) (errs *apis.FieldError) {
	if len(run.Spec.Params) == 0 {
		errs = errs.Also(apis.ErrMissingField("params"))
	}
	return errs
}

func validateExpressionsType(run *v1alpha1.Run) (errs *apis.FieldError) {
	for _, param := range run.Spec.Params {
		if param.Value.StringVal == "" {
			errs = errs.Also(apis.ErrInvalidValue(fmt.Sprintf("CEL expression parameter %s must be a string", param.Name),
				"value").ViaFieldKey("params", param.Name))
		}
	}
	return errs
}

func containsVar(varName string, params []v1beta1.Param) (bool, int) {
	for index, param := range params {
		if param.Name == varName {
			return true, index
		}
	}

	return false, -1
}

func containsParam(paramName string, vars []variablestorev1alpha1.Var) (bool, int) {
	for index, variable := range vars {
		if variable.Name == paramName {
			return true, index
		}
	}

	return false, -1
}
