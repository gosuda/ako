# ako: Opinionated Go Project Manager

[![Go Report Card](https://goreportcard.com/badge/github.com/gosuda/ako)](https://goreportcard.com/report/github.com/gosuda/ako)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`ako`는 Go 프로젝트의 생산성과 표준화를 위한 CLI 도구입니다. 반복적인 설정, 코드 구조화, Git 관리, 로컬 K3d 환경 구성을 자동화하여 개발자가 핵심 로직에 집중하도록 돕습니다.

## 해결하려는 문제 (Problem)

Go 프로젝트를 시작하고 지속적으로 관리하는 과정에는 다음과 같은 비효율과 어려움이 흔히 발생합니다.

1.  복잡하고 시간 소모적인 초기 설정:
    * 새 프로젝트마다 Go 모듈 생성, Git 저장소 초기화, 기본적인 `.gitignore` 파일 설정 등 기본적인 작업부터 시작해야 합니다.
    * CI/CD 파이프라인(예: GitHub Actions 워크플로우) 구성, 코드 품질 관리를 위한 린터(`golangci-lint`) 설정, Protobuf 사용 시 `buf` 관련 설정 등은 전문 지식이 필요하며 설정 과정이 번거롭고 오류가 발생하기 쉽습니다.
    * 개발 환경 일관성을 위한 Dev Container 설정까지 고려하면 초기 설정에 상당한 시간과 노력이 투입됩니다.

2.  일관성 없는 프로젝트 구조:
    * 명확한 가이드라인 없이는 팀이나 개인마다 `pkg`, `internal`, `cmd`, `lib` 등의 디렉토리를 다르게 해석하고 사용하게 됩니다.
    * 이는 코드의 응집도를 낮추고 의존성 관리를 복잡하게 만들며, 새로운 팀원이 프로젝트 구조를 파악하고 적응하는 데 불필요한 시간을 소모하게 합니다. 결과적으로 유지보수 비용이 증가합니다.

3.  반복적인 보일러플레이트 코드 작성:
    * 특정 레이어(예: `internal`의 HTTP 핸들러, `pkg`의 데이터베이스 클라이언트)에 필요한 기본 코드 구조, Fx 모듈 설정, 인터페이스 정의 등을 매번 유사하게 작성해야 합니다.
    * 이는 단순 반복 작업으로 개발 속도를 저하시키고, 핵심 기능 개발에 집중하는 것을 방해합니다.

4.  비효율적인 Git 워크플로우 관리:
    * 기능 개발, 버그 수정 등 목적에 맞는 브랜치를 생성하고 관리하는 규칙이 없다면 브랜치 이름이 중구난방이 되고 히스토리 추적이 어려워집니다.
    * 특히 여러 하위 작업을 포함하는 복잡한 기능 개발 시, 계층적인 브랜치 구조를 수동으로 관리하는 것은 매우 번거롭습니다.
    * 커밋 메시지 형식을 강제하지 않으면(예: Conventional Commits) 변경 사항의 의도를 파악하기 어렵고, 자동화된 변경 로그 생성이나 버전 관리에도 어려움을 겪습니다.

5.  어려운 로컬 클라우드 네이티브 환경 구축:
    * 컨테이너 기반 애플리케이션 개발 시, 로컬에서 쿠버네티스 환경을 직접 구성하는 것은 많은 노력이 필요합니다.
    * 로컬 이미지 레지스트리 설정, K3d 클러스터 생성 및 네트워크 설정, 애플리케이션 배포를 위한 매니페스트(Deployment, Service, Ingress 등) 작성 및 관리는 복잡하며 시간이 많이 소요됩니다.

## `ako`의 해결 방법 (Solution & Goals)

`ako`는 위에서 언급된 문제들을 해결하기 위해 다음과 같이 표준화되고 자동화된 기능을 제공하는 것을 목표로 합니다.

1.  원클릭 프로젝트 초기화 (`ako init` / `ako i`):
    * 단일 명령 실행만으로 Go 모듈, Git 저장소(`.gitignore` 포함), 선택 가능한 CI/CD 템플릿, `buf` 설정 및 예제, Uber Fx 의존성, Dev Container 설정, `golangci-lint` 설정 및 바이너리 설치, Conventional Commits 규칙 설정, 기본 `release` 브랜치 생성까지 프로젝트 시작에 필요한 거의 모든 것을 자동으로 구성합니다.
    * 이를 통해 개발자는 복잡한 초기 설정 과정에서 해방되어, 프로젝트 생성 즉시 핵심 코드 개발에 집중할 수 있습니다.

2.  표준화된 레이어 아키텍처 강제 및 코드 생성 (`ako go` / `ako g`):
    * `lib`, `pkg`, `internal`, `cmd` 레이어 구조를 명확히 제시하고, 각 레이어의 역할에 맞는 코드 생성을 유도합니다.
        * `lib/`: 도메인 모델, 인터페이스, DTO 등 애플리케이션의 핵심 추상화를 정의하여 의존성 방향을 관리합니다.
        * `pkg/`: 데이터베이스, 외부 API 클라이언트 등 외부 인프라와의 연동 및 구체적인 기술 구현을 담당합니다. `lib/`의 인터페이스를 구현합니다.
        * `internal/`: HTTP 핸들러, gRPC 서비스 등 애플리케이션의 사용 사례(Use Cases)와 비즈니스 로직을 구현합니다. 주로 `lib/`에 의존합니다.
        * `cmd/`: 애플리케이션의 진입점으로, 설정 로딩, Uber Fx를 이용한 의존성 주입 및 컴포넌트 조립, 애플리케이션 실행을 담당합니다.
    * `ako go internal`, `ako go pkg` 명령어는 사전 정의된 템플릿(예: Chi 핸들러, Redis 클라이언트)을 선택하여 Fx 모듈 기반의 보일러플레이트 코드를 자동으로 생성해주므로, 개발 속도를 높이고 코드의 일관성을 보장합니다.
    * `ako go buf` (`ako g f`) 명령으로 `buf generate` 실행을 간소화하여 Protobuf 기반 개발을 용이하게 합니다.

3.  체계적인 Git 워크플로우 지원 (`ako branch` / `ako b`):
    * `ako branch commit` (`ako b m`)은 Conventional Commits 규약을 따르는 커밋 메시지 작성을 대화형 프롬프트로 지원하여, 일관되고 의미 있는 Git 히스토리를 쉽게 만들 수 있도록 돕습니다.
    * `ako branch create` (`ako b c`)는 `type/scope/description` 형식의 계층적 브랜치 생성을 자동화하고, `up` (`b u`), `down` (`b d`) 명령어로 부모/자식 브랜치 간 이동을 간편하게 만들어 복잡한 기능 개발 흐름을 체계적으로 관리할 수 있게 합니다.

4.  간편한 코드 품질 검사 (`ako linter` / `ako l`):
    * `ako linter` (`ako l`) 명령 하나로 프로젝트 전체에 대해 `golangci-lint`를 실행하여, 코드 스타일 문제를 조기에 발견하고 일관된 코드 품질을 유지하도록 지원합니다.

5.  단순화된 로컬 K3d 환경 관리 (`ako k3d` / `ako k`):
    * K3d 레지스트리 및 클러스터 생성/삭제/조회 (`ako k3d registry`, `ako k3d cluster`)를 간단한 명령으로 자동화하여 로컬 쿠버네티스 인프라 구축의 복잡성을 제거합니다.
    * `ako k3d manifest init` (`ako k m i`)으로 프로젝트에 맞는 K8s 네임스페이스, 인그레스 등 기본 매니페스트 설정을 초기화합니다.
    * `ako k3d manifest create` (`ako k m c`)로 `cmd` 애플리케이션에 필요한 Deployment, Service, ConfigMap 등 매니페스트 파일을 자동으로 생성합니다.
    * `ako k3d manifest build` (`ako k m b`)로 애플리케이션의 Docker 이미지를 빌드하여 로컬 K3d 레지스트리에 푸시하고, `ako k3d manifest apply` (`ako k m a`)로 클러스터에 간편하게 배포할 수 있습니다.
    * `ako k3d manifest get` (`ako k m g`)을 통해 `kubectl get` 명령 결과를 쉽게 확인하여 배포 상태 모니터링을 용이하게 합니다.

## `ako`의 지향점 (Philosophy)

`ako`는 다음과 같은 핵심 철학을 바탕으로 설계되었습니다.

* 표준화 (Standardization): 잘 정의된 프로젝트 구조와 개발 워크플로우를 제시하여 코드의 일관성을 높이고, 팀 내외의 협업을 원활하게 하며, 장기적인 유지보수 비용을 절감하는 것을 목표로 합니다. 이는 개발자의 인지 부하를 줄여줍니다.
* 개발자 경험 (Developer Experience): 프로젝트 초기 설정, 반복적인 코드 작성, 복잡한 Git 및 K3d 관리 등 번거롭고 오류가 발생하기 쉬운 작업들을 자동화함으로써, 개발자가 실제 비즈니스 가치를 창출하는 핵심 로직 개발에 더 많은 시간과 에너지를 집중할 수 있도록 돕습니다.
* 주관성 (Opinionated): "어떻게 해야 할까?"라는 고민 대신, 레이어 기반 구조, Conventional Commits, K3d 기반 로컬 환경 등 업계에서 검증된 모범 사례들을 기반으로 명확하고 구체적인 개발 방식을 제안합니다. 이를 통해 기술적 의사결정의 피로도를 줄이고 빠른 개발 속도를 지원합니다.
* 클라우드 네이티브 준비 (Cloud-Native Ready): Protobuf를 이용한 API 설계, Docker를 통한 컨테이너화, K3d를 활용한 로컬 쿠버네티스 환경 지원을 기본으로 제공하여, 처음부터 현대적인 클라우드 환경에서의 개발 및 배포에 최적화된 프로젝트를 구성할 수 있도록 합니다.
* 효율성 (Efficiency): 개발 라이프사이클의 각 단계(초기 설정, 코드 작성, 브랜치 관리, 테스트, 배포)에서 발생하는 비효율을 제거하고 작업을 가속화하여, 전체적인 개발 생산성을 극대화하는 것을 추구합니다.

## 시작하기 (Getting Started)

### 사전 요구 사항

* [Go 1.18+](https://go.dev/dl/)
* [Git](https://git-scm.com/downloads)
* [Docker](https://docs.docker.com/get-docker/)
* [K3d](https://k3d.io/#installation) (선택 사항)
* [Buf](https://buf.build/docs/installation) (선택 사항)

### 설치

```bash
go install github.com/gosuda/ako@latest
```

### 기본 사용법 (예시)

1.  프로젝트 생성:
    ```bash
    mkdir my-project && cd my-project
    ako init # 또는 ako i
    ```
2.  코드 생성:
    ```bash
    ako go lib # 또는 ako g l (lib 레이어)
    ako go pkg # 또는 ako g p (pkg 레이어, 템플릿 선택)
    ako go internal # 또는 ako g n (internal 레이어, 템플릿 선택)
    ako go cmd # 또는 ako g c (cmd 레이어)
    ako go buf # 또는 ako g f (Protobuf 생성)
    ```
3.  Git 관리:
    ```bash
    ako branch create # 또는 ako b c (브랜치 생성)
    git add .
    ako branch commit # 또는 ako b m (커밋)
    ako branch up # 또는 ako b u (부모 브랜치 이동)
    ```
4.  린터 실행:
    ```bash
    ako linter # 또는 ako l
    ```
5.  K3d 관리:
    ```bash
    ako k3d registry create my-reg # 또는 ako k r c my-reg
    ako k3d cluster create my-clu --registry my-reg # 또는 ako k c c my-clu --registry my-reg
    ako k3d manifest init # 또는 ako k m i
    ako k3d manifest create # 또는 ako k m c (매니페스트 생성)
    ako k3d manifest build api-server # 또는 ako k m b api-server (이미지 빌드)
    ako k3d manifest apply ./deployments/manifests/api-server/*.yaml # 또는 ako k m a ...
    ako k3d manifest get pods # 또는 ako k m g p
    ```

## 명령어 단축키 (Command Aliases)

자주 사용하는 명령어와 가장 짧은 단축키 목록입니다.

* `ako init` -> `ako i`
* `ako go lib` -> `ako g l`
* `ako go pkg` -> `ako g p`
* `ako go internal` -> `ako g n`
* `ako go cmd` -> `ako g c`
* `ako go buf` -> `ako g f`
* `ako branch current` -> `ako b n`
* `ako branch commit` -> `ako b m`
* `ako branch create` -> `ako b c`
* `ako branch up` -> `ako b u`
* `ako branch down` -> `ako b d`
* `ako linter` -> `ako l`
* `ako k3d registry list` -> `ako k r l` / `ls`
* `ako k3d registry create` -> `ako k r c`
* `ako k3d registry delete` -> `ako k r d` / `rm`
* `ako k3d cluster list` -> `ako k c l` / `ls`
* `ako k3d cluster create` -> `ako k c c`
* `ako k3d cluster delete` -> `ako k c d` / `rm`
* `ako k3d cluster append-port` -> `ako k c a` / `ap`
* `ako k3d manifest init` -> `ako k m i` / `f i`
* `ako k3d manifest create` -> `ako k m c` / `f c`
* `ako k3d manifest build` -> `ako k m b` / `f b` / `f d`
* `ako k3d manifest apply` -> `ako k m a` / `f a`
* `ako k3d manifest get pods` -> `ako k m g p` / `f g p` / `f g po`
* `ako k3d manifest get services` -> `ako k m g s` / `f g s` / `f g svc`
* `ako k3d manifest get deployments` -> `ako k m g d` / `f g d` / `f g deploy`
* `ako k3d manifest get ingress` -> `ako k m g i` / `f g i`

## 기반 기술 (Core Technologies)

* Go: 핵심 개발 언어.
* urfave/cli/v3 & AlecAivazis/survey/v2: CLI 인터페이스 구축.
* Git: 버전 관리 및 Conventional Commits 기반 워크플로우 자동화.
* Buf: Protobuf 스키마 관리 및 코드 생성.
* golangci-lint: 코드 정적 분석.
* Docker: 컨테이너화 및 이미지 빌드 자동화.
* K3d: 로컬 쿠버네티스 환경 구성 자동화.
* Uber Fx: 의존성 주입 프레임워크 (템플릿 기반).

## 지원하는 코드 템플릿 (Supported Code Templates)

`ako go internal` 및 `ako go pkg` 실행 시 선택 가능한 Fx 기반 템플릿입니다.

### Internal Layer Templates (`ako g n`)

* `chi`: go-chi/chi 라우터 기반 핸들러.
* `fiber`: gofiber/fiber 프레임워크 기반 핸들러.
* `grpc_server`: gRPC 서버 구현.
* `empty`: 기본 Fx 모듈 구조만 포함.

### Pkg Layer Templates (`ako g p`)

* `cassandra`: gocql/gocql 클라이언트.
* `clickhouse`: ClickHouse/clickhouse-go 클라이언트.
* `duckdb`: marcboeker/go-duckdb 클라이언트.
* `entgo`: ent/ent ORM 설정.
* `grpc_client`: gRPC 클라이언트.
* `http_client`: 표준 `net/http` 클라이언트.
* `kafka`: IBM/sarama 프로듀서/컨슈머.
* `minio`: minio/minio-go 클라이언트.
* `mssql`: microsoft/go-mssqldb 클라이언트.
* `nats`: nats-io/nats.go 클라이언트.
* `qdrant`: qdrant/go-client 클라이언트.
* `redis`: redis/rueidis 클라이언트.
* `sqlc`: sqlc-dev/sqlc 설정 (SQL DB용).
* `templ`: a-h/templ SSR 컴포넌트.
* `valkey`: valkey-io/valkey-go 클라이언트.
* `vault`: hashicorp/vault/api 클라이언트.
* `empty`: 기본 Fx 모듈 구조만 포함.

## 라이선스 (License)

MIT License. 자세한 내용은 `LICENSE` 파일을 참고해주세요.
