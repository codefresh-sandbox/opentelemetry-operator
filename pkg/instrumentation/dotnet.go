// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package instrumentation

import (
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/open-telemetry/opentelemetry-operator/apis/v1alpha1"
)

const (
	envDotNetCoreClrEnableProfiling     = "CORECLR_ENABLE_PROFILING"
	envDotNetCoreClrProfiler            = "CORECLR_PROFILER"
	envDotNetCoreClrProfilerPath        = "CORECLR_PROFILER_PATH"
	envDotNetAdditionalDeps             = "DOTNET_ADDITIONAL_DEPS"
	envDotNetSharedStore                = "DOTNET_SHARED_STORE"
	envDotNetStartupHook                = "DOTNET_STARTUP_HOOKS"
	envDotNetOTelAutoHome               = "OTEL_DOTNET_AUTO_HOME"
	dotNetCoreClrEnableProfilingEnabled = "1"
	dotNetCoreClrProfilerId             = "{918728DD-259F-4A6A-AC2B-B85E1B658318}"
	dotNetCoreClrProfilerPath           = "/otel-auto-instrumentation/OpenTelemetry.AutoInstrumentation.Native.so"
	dotNetAdditionalDepsPath            = "/otel-auto-instrumentation/AdditionalDeps"
	dotNetOTelAutoHomePath              = "/otel-auto-instrumentation"
	dotNetSharedStorePath               = "/otel-auto-instrumentation/store"
	dotNetStartupHookPath               = "/otel-auto-instrumentation/net/OpenTelemetry.AutoInstrumentation.StartupHook.dll"
)

func injectDotNetSDK(dotNetSpec v1alpha1.DotNet, pod corev1.Pod, index int) (corev1.Pod, error) {

	// caller checks if there is at least one container.
	container := &pod.Spec.Containers[index]

	err := validateContainerEnv(container.Env, envDotNetStartupHook, envDotNetAdditionalDeps, envDotNetSharedStore)
	if err != nil {
		return pod, err
	}

	// check if OTEL_DOTNET_AUTO_HOME env var is already set in the container
	// if it is already set, then we assume that .NET Auto-instrumentation is already configured for this container
	if getIndexOfEnv(container.Env, envDotNetOTelAutoHome) > -1 {
		return pod, errors.New("OTEL_DOTNET_AUTO_HOME environment variable is already set in the container")
	}

	// check if OTEL_DOTNET_AUTO_HOME env var is already set in the .NET instrumentatiom spec
	// if it is already set, then we assume that .NET Auto-instrumentation is already configured for this container
	if getIndexOfEnv(dotNetSpec.Env, envDotNetOTelAutoHome) > -1 {
		return pod, errors.New("OTEL_DOTNET_AUTO_HOME environment variable is already set in the .NET instrumentation spec")
	}

	// inject .NET instrumentation spec env vars.
	for _, env := range dotNetSpec.Env {
		idx := getIndexOfEnv(container.Env, env.Name)
		if idx == -1 {
			container.Env = append(container.Env, env)
		}
	}

	const (
		doNotConcatEnvValues = false
		concatEnvValues      = true
	)

	setDotNetEnvVar(container, envDotNetCoreClrEnableProfiling, dotNetCoreClrEnableProfilingEnabled, doNotConcatEnvValues)

	setDotNetEnvVar(container, envDotNetCoreClrProfiler, dotNetCoreClrProfilerId, doNotConcatEnvValues)

	setDotNetEnvVar(container, envDotNetCoreClrProfilerPath, dotNetCoreClrProfilerPath, doNotConcatEnvValues)

	setDotNetEnvVar(container, envDotNetStartupHook, dotNetStartupHookPath, concatEnvValues)

	setDotNetEnvVar(container, envDotNetAdditionalDeps, dotNetAdditionalDepsPath, concatEnvValues)

	setDotNetEnvVar(container, envDotNetOTelAutoHome, dotNetOTelAutoHomePath, doNotConcatEnvValues)

	setDotNetEnvVar(container, envDotNetSharedStore, dotNetSharedStorePath, concatEnvValues)

	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
		Name:      volumeName,
		MountPath: "/otel-auto-instrumentation",
	})

	// We just inject Volumes and init containers for the first processed container.
	if isInitContainerMissing(pod) {
		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			}})

		pod.Spec.InitContainers = append(pod.Spec.InitContainers, corev1.Container{
			Name:    initContainerName,
			Image:   dotNetSpec.Image,
			Command: []string{"cp", "-a", "/autoinstrumentation/.", "/otel-auto-instrumentation/"},
			VolumeMounts: []corev1.VolumeMount{{
				Name:      volumeName,
				MountPath: "/otel-auto-instrumentation",
			}},
		})
	}
	return pod, nil
}

// setDotNetEnvVar function sets env var to the container if not exist already.
// value of concatValues should be set to true if the env var supports multiple values separated by :.
// If it is set to false, the original container's env var value has priority.
func setDotNetEnvVar(container *corev1.Container, envVarName string, envVarValue string, concatValues bool) {
	idx := getIndexOfEnv(container.Env, envVarName)
	if idx < 0 {
		container.Env = append(container.Env, corev1.EnvVar{
			Name:  envVarName,
			Value: envVarValue,
		})
		return
	}
	if concatValues {
		container.Env[idx].Value = fmt.Sprintf("%s:%s", container.Env[idx].Value, envVarValue)
	}
}
