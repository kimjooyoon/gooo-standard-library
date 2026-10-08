# Gooo standard library — experimental

Gooo 프로그램에서 가져다 쓰는 작은 함수 모음입니다. 숫자·논리·문자열 함수
17개를 Gooo로 작성했습니다. 함수가 같은 패키지의 다른 함수를 부르고,
앱이 여러 패키지를 가져와 함께 쓰는 흐름까지 실행합니다.

함수의 동작은 `library/*.gooo`에 있습니다. Go 코드는 입력 전달과 결과 비교,
관측 파일 저장을 맡습니다. Go 1.27.1과
[Gooo 0.6.10 개발판](https://github.com/kimjooyoon/meta-ontology-go/releases/tag/v0.6.10-dev)을 사용합니다.

## 들어 있는 함수

| 패키지 | 함수 |
| --- | --- |
| `std/numbers` | `Min`, `Max`, `Clamp`, `AbsSaturating`, `Sign`, `IsZero`, `InRange`, `NonNegative` |
| `std/logic` | `And`, `Or`, `Not` |
| `std/text` | `Coalesce`, `Choose`, `HasPrefix`, `HasSuffix`, `StripSuffix`, `ByteLength` |

`Clamp`와 `InRange`는 뒤집힌 두 경계를 정렬합니다. int64 최솟값의 절댓값은
최댓값으로 포화시킵니다. 문자열 길이는 UTF-8 바이트 수이며 현재 Text 입력의
한도는 1,024바이트입니다. 빈 접두사와 접미사는 일치합니다.

## 바로 실행하기

예제는 제목과 예산을 받아 사용 여부, 빈 제목의 대체값, 사용할 바이트 수를
계산합니다. 세 패키지의 함수를 부르고 결과의 세 필드를 조립합니다.

```sh
gooo package execute --json --inputs examples/preview-inputs.json gooo.workspace.json > execution.json
gooo package replay --json --receipt execution.json --inputs examples/preview-inputs.json gooo.workspace.json
```

입력만 주는 사용 경로라 실행 상태는 `OBSERVED`입니다. 조립에 쓴 다섯 예시의
결과와 실제 사용 입력에 대한 정답률을 따로 읽습니다. 저장 재실행에는 모델이 필요하지 않습니다.

모델을 사용하려면 첫 명령에 `--assembly-model /absolute/path/to/model.json`을
추가합니다. 이번 관측은 [공개 그래프 QAT 모델](https://github.com/kimjooyoon/gooo-ecosystem-workbench/tree/19485bb64278ab8ac219e042f92af677ff3f5051/models/graph-chooser-20261008/all-data-demonstration/qat_ternary)을
재학습 없이 사용했습니다. 기본 실행은 고정된 후보 순서입니다.

## 자신의 프로그램에서 쓰기

```gooo
package budget
namespace budget
import numbers "std/numbers"
entity Integer id "gooo://std/integer"
activity Bound(Integer, Integer) -> Integer computes `return numbers.Clamp(input0, 0, input1)`
```

사용할 `.gooo` 파일을 작업공간에 넣고 `gooo.workspace.json`의 패키지·소스·가져오기
목록에 등록합니다. 이 저장소의 [작업공간 명세](gooo.workspace.json)와
[호출하는 예제](examples/preview.gooo)가 전체 구성입니다. 현재 배포 단위는
Git으로 버전을 고정한 소스 파일과 작업공간 명세입니다.

## 확인한 범위

함수 17개에 명시한 입력 79건을 실행하고 저장 재실행했습니다. 가져다 쓰는 앱은
한 개이며, 조립용 다섯 예시와 겹치지 않는 별도 입력 열 개를 실행했습니다.

| 후보 한도 | 고정 순서: 입력 / 필드 | 자체 모델: 입력 / 필드 |
| --- | --- | --- |
| 1 | 1/10 · 14/30 | 1/10 · 14/30 |
| 2 | 1/10 · 20/30 | 3/10 · 17/30 |
| 4 | 3/10 · 23/30 | 4/10 · 24/30 |
| 8 | 10/10 · 30/30 | 10/10 · 30/30 |

이번 앱에서는 두 방식 모두 여덟 후보를 시도해야 완성됐습니다. 적은 한도에서
맞힌 입력 수와 필드 수는 서로 다른 모습을 보였습니다. 각 부분 결과도 그대로
재실행됐습니다. `PLAN.md`와 함수·평가 입력을 먼저 고정한 뒤 모델을 호출했습니다.
계획·원본·실행 환경은 [관측 기록](publication/initial/README.md)에 있습니다.

전체 검사를 직접 실행하려면:

```sh
go run ./cmd/verify --compiler /path/to/gooo --out out/fixed
go run ./cmd/verify --compiler /path/to/gooo --model /path/to/model.json --out out/paired
```

각각 새 출력 폴더를 사용합니다. 관측기는 int64를 보존해 실제 값과 기대값을
비교하고, 입력·출력 활동·재실행·새 모델 호출 수를 확인합니다.
CI도 공개된 컴파일러 파일과 고정한 모델 소스로 같은 예제를 실행합니다.

기존 [생태계 작업장](https://github.com/kimjooyoon/gooo-ecosystem-workbench)의 기본 함수와
컴파일러의 문자열 예제에서 출발했습니다. 작은 도구에서 같은 판단과 계산을
가져다 쓰면서, Gooo의 패키지와 함수 표현을 발전시키려 합니다. MIT 라이선스입니다.
