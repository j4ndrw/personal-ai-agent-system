import importlib.util
import inspect
import os
from dataclasses import asdict, dataclass, field
from functools import wraps
from typing import Any, Callable, Literal

from ollama import Message

ToolRepository = dict[str, Callable]
ToolHandlers = dict[Callable, Callable[[Message.ToolCall], list[Any]]]


@dataclass
class Toolkit:
    repository: ToolRepository
    handlers: ToolHandlers


@dataclass
class _Description:
    details: str
    args: list[tuple[str, str]] = field(default_factory=lambda: [])
    returns: list[tuple[str, str]] = field(default_factory=lambda: [])


@dataclass
class _Error:
    function_kind: Literal["tool"] | Literal["resource"]
    function_origin: str
    error: str

    def dict(self):
        return {k: v for k, v in asdict(self).items()}


@dataclass
class _ToolSuccess:
    function_origin: str
    success: bool = field(default_factory=lambda: True)

    def dict(self):
        return {k: v for k, v in asdict(self).items()}


def create_tool_repository(*functions: Callable) -> ToolRepository:
    return {function.__name__: function for function in functions}


def create_tool_handlers() -> ToolHandlers:
    return {}


def create_toolkit():
    return Toolkit(repository=create_tool_repository(), handlers=create_tool_handlers())


def _update_toolkit(
    *,
    toolkit: Toolkit,
):
    def decorator(func: Callable):
        @wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        toolkit.repository = {**toolkit.repository, **create_tool_repository(wrapper)}
        toolkit.handlers = {
            **toolkit.handlers,
            wrapper: lambda tool_call: [
                *map(
                    lambda arg: tool_call.function.arguments.get(arg),
                    [*inspect.signature(wrapper).parameters.keys()],
                )
            ],
        }

        return wrapper

    return decorator


def _update_function_docstring(
    *,
    kind: Literal["tool"] | Literal["resource"],
    description: _Description | None = None,
):
    def decorator(func: Callable):
        @wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        if description is None:
            return wrapper

        wrapper.__doc__ = f"\n    Function type: {kind.capitalize()}"
        wrapper.__doc__ = f"\n    ---"
        wrapper.__doc__ = f"{description.details}"
        if len(description.args) > 0:
            wrapper.__doc__ += "\n\n    Args: \n        "
            wrapper.__doc__ += "\n        ".join(
                [*map(lambda arg: f"{arg[0]}: {arg[1]}", description.args)]
            )
        if len(description.returns) > 0:
            wrapper.__doc__ += "\n\n    Returns: \n        "
            wrapper.__doc__ += "\n        ".join(
                [*map(lambda ret: f"{ret[0]}: {ret[1]}", description.returns)]
            )

        return wrapper

    return decorator


def define_toolkit():
    toolkit = create_toolkit()
    register_toolkit = lambda: toolkit

    class tool:
        @staticmethod
        def create(
            *,
            description: _Description | None = None,
        ):
            def decorator(func: Callable):
                @_update_toolkit(toolkit=toolkit)
                @_update_function_docstring(kind="tool", description=description)
                @wraps(func)
                def wrapper(*args, **kwargs):
                    return func(*args, **kwargs)

                return wrapper

            return decorator

        @staticmethod
        def error(function_origin: str, *, error: str) -> dict:
            return _Error(
                function_kind="tool", function_origin=function_origin, error=error
            ).dict()

        @staticmethod
        def success(function_origin: str) -> dict:
            return _ToolSuccess(function_origin=function_origin).dict()

    class resource:
        @staticmethod
        def create(
            *,
            description: _Description | None = None,
        ):
            def decorator(func: Callable):
                @_update_toolkit(toolkit=toolkit)
                @_update_function_docstring(kind="resource", description=description)
                @wraps(func)
                def wrapper(*args, **kwargs):
                    return func(*args, **kwargs)

                return wrapper

            return decorator

        @staticmethod
        def error(function_origin: str, *, error: str) -> dict:
            return _Error(
                function_kind="resource", function_origin=function_origin, error=error
            ).dict()

    return tool, resource, register_toolkit


class tool:
    @staticmethod
    def error(function_origin: str, *, error: str) -> dict:
        return _Error(
            function_kind="tool", function_origin=function_origin, error=error
        ).dict()

    @staticmethod
    def success(function_origin: str) -> dict:
        return _ToolSuccess(function_origin=function_origin).dict()


class resource:
    @staticmethod
    def error(function_origin: str, *, error: str) -> dict:
        return _Error(
            function_kind="resource", function_origin=function_origin, error=error
        ).dict()


def description(
    details: str,
    *,
    args: list[tuple[str, str]] | None = None,
    returns: list[tuple[str, str]] | None = None,
) -> _Description:
    return _Description(details=details, args=args or [], returns=returns or [])


def load_toolkits(path: str, names: list[str]) -> list[Toolkit]:
    toolkits: list[Toolkit] = []
    for dirpath, _, filenames in os.walk(os.path.abspath(path)):
        for filename in filenames:
            if filename.endswith(".py") and filename != "__init__.py":
                module_name = filename[:-3]
                if filename in names:
                    module_path = os.path.join(dirpath, filename)

                    package_name = os.path.relpath(
                        dirpath, start=os.path.curdir
                    ).replace(os.path.sep, ".")
                    full_module_name = (
                        f"{package_name}.{module_name}" if package_name else module_name
                    )
                    spec = importlib.util.spec_from_file_location(
                        full_module_name, module_path
                    )
                    if spec is not None and spec.loader is not None:
                        module = importlib.util.module_from_spec(spec)
                        spec.loader.exec_module(module)

                        if hasattr(module, "register_toolkit"):
                            toolkits.append(module.register_toolkit())

    return toolkits
