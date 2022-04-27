"""Describe methods to utilize the VS/VSR resources."""

import logging

import yaml
from kubernetes.client import CoreV1Api, CustomObjectsApi
from kubernetes.client.rest import ApiException
from suite.custom_resources_utils import read_custom_resource
from suite.resources_utils import ensure_item_removal, get_file_contents


def read_vs(custom_objects: CustomObjectsApi, namespace, name) -> object:
    """
    Read VirtualServer resource.
    """
    return read_custom_resource(custom_objects, namespace, "virtualservers", name)


def read_vsr(custom_objects: CustomObjectsApi, namespace, name) -> object:
    """
    Read VirtualServerRoute resource.
    """
    return read_custom_resource(custom_objects, namespace, "virtualserverroutes", name)


def create_virtual_server_from_yaml(
    custom_objects: CustomObjectsApi, yaml_manifest, namespace
) -> str:
    """
    Create a VirtualServer based on yaml file.

    :param custom_objects: CustomObjectsApi
    :param yaml_manifest: an absolute path to file
    :param namespace:
    :return: str
    """
    print("Create a VirtualServer:")
    with open(yaml_manifest) as f:
        dep = yaml.safe_load(f)

    return create_virtual_server(custom_objects, dep, namespace)


def create_virtual_server(custom_objects: CustomObjectsApi, vs, namespace) -> str:
    """
    Create a VirtualServer.

    :param custom_objects: CustomObjectsApi
    :param vs: a VirtualServer
    :param namespace:
    :return: str
    """
    print("Create a VirtualServer:")
    try:
        custom_objects.create_namespaced_custom_object(
            "k8s.nginx.org", "v1", namespace, "virtualservers", vs
        )
        print(f"VirtualServer created with name '{vs['metadata']['name']}'")
        return vs["metadata"]["name"]
    except ApiException as ex:
        logging.exception(
            f"Exception: {ex} occurred while creating VirtualServer: {vs['metadata']['name']}"
        )
        raise


def delete_virtual_server(custom_objects: CustomObjectsApi, name, namespace) -> None:
    """
    Delete a VirtualServer.

    :param custom_objects: CustomObjectsApi
    :param namespace: namespace
    :param name:
    :return:
    """
    print(f"Delete a VirtualServer: {name}")

    custom_objects.delete_namespaced_custom_object(
        "k8s.nginx.org", "v1", namespace, "virtualservers", name
    )
    ensure_item_removal(
        custom_objects.get_namespaced_custom_object,
        "k8s.nginx.org",
        "v1",
        namespace,
        "virtualservers",
        name,
    )
    print(f"VirtualServer was removed with name '{name}'")


def patch_virtual_server_from_yaml(
    custom_objects: CustomObjectsApi, name, yaml_manifest, namespace
) -> None:
    """
    Patch a VS based on yaml manifest
    :param custom_objects: CustomObjectsApi
    :param name:
    :param yaml_manifest: an absolute path to file
    :param namespace:
    :return:
    """
    print(f"Update a VirtualServer: {name}, namespace: {namespace}")
    with open(yaml_manifest) as f:
        dep = yaml.safe_load(f)

    try:
        print(f"Try to patch VirtualServer: {dep}")
        custom_objects.patch_namespaced_custom_object(
            "k8s.nginx.org", "v1", namespace, "virtualservers", name, dep
        )
        print(f"VirtualServer updated with name '{dep['metadata']['name']}'")
    except ApiException:
        logging.exception(f"Failed with exception while patching VirtualServer: {name}")
        raise
    except Exception as ex:
        logging.exception(f"Failed with exception while patching VirtualServer: {name}, Exception: {ex.with_traceback}")
        raise


def delete_and_create_vs_from_yaml(
    custom_objects: CustomObjectsApi, name, yaml_manifest, namespace
) -> None:
    """
    Perform delete and create for vs with same name based on yaml

    :param custom_objects: CustomObjectsApi
    :param name:
    :param yaml_manifest: an absolute path to file
    :param namespace:
    :return:
    """
    try:
        delete_virtual_server(custom_objects, name, namespace)
        create_virtual_server_from_yaml(custom_objects, yaml_manifest, namespace)
    except ApiException:
        logging.exception(f"Failed with exception while patching VirtualServer: {name}")
        raise


def patch_virtual_server(custom_objects: CustomObjectsApi, name, namespace, body) -> str:
    """
    Update a VirtualServer based on a dict.

    :param custom_objects: CustomObjectsApi
    :param name:
    :param body: dict
    :param namespace:
    :return: str
    """
    print("Update a VirtualServer:")
    custom_objects.patch_namespaced_custom_object(
        "k8s.nginx.org", "v1", namespace, "virtualservers", name, body
    )
    print(f"VirtualServer updated with a name '{body['metadata']['name']}'")
    return body["metadata"]["name"]


def patch_v_s_route_from_yaml(
    custom_objects: CustomObjectsApi, name, yaml_manifest, namespace
) -> None:
    """
    Update a VirtualServerRoute based on yaml manifest

    :param custom_objects: CustomObjectsApi
    :param name:
    :param yaml_manifest: an absolute path to file
    :param namespace:
    :return:
    """
    print(f"Update a VirtualServerRoute: {name}, namespace: {namespace}")
    with open(yaml_manifest) as f:
        dep = yaml.safe_load(f)
    try:
        print(f"Try to patch VirtualServerRoute: {dep}")
        custom_objects.patch_namespaced_custom_object(
            "k8s.nginx.org", "v1", namespace, "virtualserverroutes", name, dep
        )
        print(f"VirtualServerRoute updated with name '{dep['metadata']['name']}'")
    except ApiException:
        logging.exception(f"Failed with exception while patching VirtualServerRoute: {name}")
        raise
    except Exception as ex:
        logging.exception(
            f"Failed with exception while patching VirtualServerRoute: {name}, Exception: {ex.with_traceback}")
        raise


def get_vs_nginx_template_conf(
    v1: CoreV1Api, vs_namespace, vs_name, pod_name, pod_namespace
) -> str:
    """
    Get contents of /etc/nginx/conf.d/vs_{namespace}_{vs_name}.conf in the pod.

    :param v1: CoreV1Api
    :param vs_namespace:
    :param vs_name:
    :param pod_name:
    :param pod_namespace:
    :return: str
    """
    file_path = f"/etc/nginx/conf.d/vs_{vs_namespace}_{vs_name}.conf"
    return get_file_contents(v1, file_path, pod_name, pod_namespace)


def create_v_s_route_from_yaml(custom_objects: CustomObjectsApi, yaml_manifest, namespace) -> str:
    """
    Create a VirtualServerRoute based on a yaml file.

    :param custom_objects: CustomObjectsApi
    :param yaml_manifest: an absolute path to a file
    :param namespace:
    :return: str
    """
    print("Create a VirtualServerRoute:")
    with open(yaml_manifest) as f:
        dep = yaml.safe_load(f)

    return create_v_s_route(custom_objects, dep, namespace)


def create_v_s_route(custom_objects: CustomObjectsApi, vsr, namespace) -> str:
    """
    Create a VirtualServerRoute.

    :param custom_objects: CustomObjectsApi
    :param vsr: a VirtualServerRoute
    :param namespace:
    :return: str
    """
    print("Create a VirtualServerRoute:")
    custom_objects.create_namespaced_custom_object(
        "k8s.nginx.org", "v1", namespace, "virtualserverroutes", vsr
    )
    print(f"VirtualServerRoute created with a name '{vsr['metadata']['name']}'")
    return vsr["metadata"]["name"]


def patch_v_s_route(custom_objects: CustomObjectsApi, name, namespace, body) -> str:
    """
    Update a VirtualServerRoute based on a dict.

    :param custom_objects: CustomObjectsApi
    :param name:
    :param body: dict
    :param namespace:
    :return: str
    """
    print("Update a VirtualServerRoute:")
    custom_objects.patch_namespaced_custom_object(
        "k8s.nginx.org", "v1", namespace, "virtualserverroutes", name, body
    )
    print(f"VirtualServerRoute updated with a name '{body['metadata']['name']}'")
    return body["metadata"]["name"]


def delete_v_s_route(custom_objects: CustomObjectsApi, name, namespace) -> None:
    """
    Delete a VirtualServerRoute.

    :param custom_objects: CustomObjectsApi
    :param namespace: namespace
    :param name:
    :return:
    """
    print(f"Delete a VirtualServerRoute: {name}")
    custom_objects.delete_namespaced_custom_object(
        "k8s.nginx.org",
        "v1",
        namespace,
        "virtualserverroutes",
        name,
    )
    ensure_item_removal(
        custom_objects.get_namespaced_custom_object,
        "k8s.nginx.org",
        "v1",
        namespace,
        "virtualserverroutes",
        name,
    )
    print(f"VirtualServerRoute was removed with the name '{name}'")
