// +build !ignore_autogenerated

/*
Copyright The Kubernetes Authors.

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RhpamBusinessCentralConfig) DeepCopyInto(out *RhpamBusinessCentralConfig) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RhpamBusinessCentralConfig.
func (in *RhpamBusinessCentralConfig) DeepCopy() *RhpamBusinessCentralConfig {
	if in == nil {
		return nil
	}
	out := new(RhpamBusinessCentralConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RhpamConfig) DeepCopyInto(out *RhpamConfig) {
	*out = *in
	out.DatabaseConfig = in.DatabaseConfig
	out.BusinessCentralConfig = in.BusinessCentralConfig
	out.KieServerConfig = in.KieServerConfig
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RhpamConfig.
func (in *RhpamConfig) DeepCopy() *RhpamConfig {
	if in == nil {
		return nil
	}
	out := new(RhpamConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RhpamDatabaseConfig) DeepCopyInto(out *RhpamDatabaseConfig) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RhpamDatabaseConfig.
func (in *RhpamDatabaseConfig) DeepCopy() *RhpamDatabaseConfig {
	if in == nil {
		return nil
	}
	out := new(RhpamDatabaseConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RhpamDev) DeepCopyInto(out *RhpamDev) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RhpamDev.
func (in *RhpamDev) DeepCopy() *RhpamDev {
	if in == nil {
		return nil
	}
	out := new(RhpamDev)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RhpamDev) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RhpamDevList) DeepCopyInto(out *RhpamDevList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RhpamDev, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RhpamDevList.
func (in *RhpamDevList) DeepCopy() *RhpamDevList {
	if in == nil {
		return nil
	}
	out := new(RhpamDevList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RhpamDevList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RhpamDevSpec) DeepCopyInto(out *RhpamDevSpec) {
	*out = *in
	out.Config = in.Config
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RhpamDevSpec.
func (in *RhpamDevSpec) DeepCopy() *RhpamDevSpec {
	if in == nil {
		return nil
	}
	out := new(RhpamDevSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RhpamDevStatus) DeepCopyInto(out *RhpamDevStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RhpamDevStatus.
func (in *RhpamDevStatus) DeepCopy() *RhpamDevStatus {
	if in == nil {
		return nil
	}
	out := new(RhpamDevStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RhpamKieServerConfig) DeepCopyInto(out *RhpamKieServerConfig) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RhpamKieServerConfig.
func (in *RhpamKieServerConfig) DeepCopy() *RhpamKieServerConfig {
	if in == nil {
		return nil
	}
	out := new(RhpamKieServerConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RhpamUser) DeepCopyInto(out *RhpamUser) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RhpamUser.
func (in *RhpamUser) DeepCopy() *RhpamUser {
	if in == nil {
		return nil
	}
	out := new(RhpamUser)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RhpamUser) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RhpamUserList) DeepCopyInto(out *RhpamUserList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RhpamUser, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RhpamUserList.
func (in *RhpamUserList) DeepCopy() *RhpamUserList {
	if in == nil {
		return nil
	}
	out := new(RhpamUserList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RhpamUserList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RhpamUserSpec) DeepCopyInto(out *RhpamUserSpec) {
	*out = *in
	if in.Roles != nil {
		in, out := &in.Roles, &out.Roles
		*out = make([]*Role, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Role)
				**out = **in
			}
		}
	}
	if in.Users != nil {
		in, out := &in.Users, &out.Users
		*out = make([]*User, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(User)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RhpamUserSpec.
func (in *RhpamUserSpec) DeepCopy() *RhpamUserSpec {
	if in == nil {
		return nil
	}
	out := new(RhpamUserSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RhpamUserStatus) DeepCopyInto(out *RhpamUserStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RhpamUserStatus.
func (in *RhpamUserStatus) DeepCopy() *RhpamUserStatus {
	if in == nil {
		return nil
	}
	out := new(RhpamUserStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Role) DeepCopyInto(out *Role) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Role.
func (in *Role) DeepCopy() *Role {
	if in == nil {
		return nil
	}
	out := new(Role)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *User) DeepCopyInto(out *User) {
	*out = *in
	if in.Roles != nil {
		in, out := &in.Roles, &out.Roles
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new User.
func (in *User) DeepCopy() *User {
	if in == nil {
		return nil
	}
	out := new(User)
	in.DeepCopyInto(out)
	return out
}
