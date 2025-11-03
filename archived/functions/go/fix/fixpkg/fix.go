package fixpkg

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	v1 "github.com/kptdev/krm-functions-catalog/archived/functions/go/fix/v1"
	"github.com/kptdev/krm-functions-catalog/archived/functions/go/fix/v1alpha1"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/sets"
	"sigs.k8s.io/kustomize/kyaml/setters2"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ kio.Filter = &Fix{}

// Fix migrates the resources in v1alpha1 format to v1 format
type Fix struct {
	// pkgPathToPkgFilePaths key: package path relative to root package path
	// value: list of file paths in the package, relative to root package
	// the list doesn't include the file paths of subpackages
	pkgPathToPkgFilePaths map[string]sets.String

	// pkgFileToPkgPath key: file path(relative to root package)
	// value: path to package(relative to root package) to which the file belongs to
	pkgFileToPkgPath map[string]string

	// settersSchema is the schema equivalent of openAPI section in Kptfile
	// this must be updated while processing each resources so that the visitor
	// interface methods have access to it
	settersSchema *spec.Schema

	// settersConfigs holds the newly created setter configs as part of migration
	settersConfigs []*yaml.RNode

	// Results are the results of fixing packages
	Results []*Result
}

// Result holds result of fixing packages
type Result struct {
	// FilePath is the file path of the resource
	FilePath string

	// Message is the result message
	Message string
}

const SettersConfigFileName = "setters.yaml"

// Filter implements Fix as a yaml.Filter
func (s *Fix) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	// group the resources based on the packages they belong to and
	// populate Fix struct maps
	if err := s.groupPathsInPkgs(nodes); err != nil {
		return nodes, fmt.Errorf("unable to group resources by packages, %q", err.Error())
	}

	// each kpt package has a Kptfile with OpenAPI section(could be empty)
	// get the map of pkgPath to OpenAPI schema as a pre-processing step
	pkgPathToSettersSchema, err := getPkgPathToSettersSchema(nodes)
	if err != nil {
		return nodes, err
	}

	kfFound := false
	for i := range nodes {
		meta, err := nodes[i].GetMeta()
		if err != nil {
			return nodes, err
		}

		// check if there is Kptfile in root package
		if meta.Kind == v1.KptFileName {
			kfFound = true
		}

		if meta.Kind == v1.KptFileName {
			// this node is Kptfile node
			// migrate Kptfile from v1alpha1 to v1
			pkgPath := filepath.Dir(meta.Annotations[kioutil.PathAnnotation])
			functions := s.FunctionsInPkg(nodes, pkgPath)
			kNode, err := s.FixKptfile(nodes[i], functions)
			if err != nil {
				return nodes, err
			}
			nodes[i] = kNode
			continue
		}

		if meta.Labels["cli-utils.sigs.k8s.io/inventory-id"] != "" {
			s.Results = append(s.Results, &Result{
				FilePath: meta.Annotations[kioutil.PathAnnotation],
				Message:  `Please refer to https://kpt.dev/reference/cli/live/migrate/, this package is using "inventory-object"`,
			})
			continue
		}

		// this is not a Kptfile node
		// get the setters schema of the package to which the resource belongs
		pkgPathOfResource := s.pkgFileToPkgPath[meta.Annotations[kioutil.PathAnnotation]]

		// update s.settersSchema so that visitor interface has setters schema for resource
		s.settersSchema = pkgPathToSettersSchema[pkgPathOfResource]

		// fix setter comments in each resource
		err = accept(s, nodes[i], s.settersSchema)
		if err != nil {
			return nil, errors.Wrap(err)
		}
	}

	if !kfFound {
		return nodes, fmt.Errorf("kptfile not found in directory tree, make sure you specify '--include-meta-resources' flag")
	}

	// add all the newly created setter configs
	nodes = append(nodes, s.settersConfigs...)

	return nodes, nil
}

// groupPathsInPkgs takes the input nodes list and populates
// pkgPathToPkgFilePaths and pkgFileToPkgPath in Fix struct
// this is a one-time pre-processing step to group package paths
func (s *Fix) groupPathsInPkgs(nodes []*yaml.RNode) error {
	s.pkgPathToPkgFilePaths = make(map[string]sets.String)
	s.pkgFileToPkgPath = make(map[string]string)
	nonKfPaths, err := getNonKptFilesPaths(nodes)
	if err != nil {
		return err
	}
	kfPaths, err := getKptFilesPaths(nodes)
	if err != nil {
		return err
	}
	for _, kfPath := range kfPaths.List() {
		filesInPkg := filesInPackage(filepath.Dir(kfPath), nonKfPaths, kfPaths)
		s.pkgPathToPkgFilePaths[filepath.Dir(kfPath)] = filesInPkg
		for _, filePath := range filesInPkg.List() {
			s.pkgFileToPkgPath[filePath] = filepath.Dir(kfPath)
		}
	}
	return nil
}

// getNonKptFilesPaths returns the file paths of all resources which are NOT Kptfiles
func getNonKptFilesPaths(nodes []*yaml.RNode) (sets.String, error) {
	paths := sets.String{}
	for _, node := range nodes {
		meta, err := node.GetMeta()
		if err != nil {
			return nil, err
		}
		path := meta.Annotations[kioutil.PathAnnotation]
		if filepath.Base(path) != v1.KptFileName {
			paths.Insert(meta.Annotations[kioutil.PathAnnotation])
		}
	}
	return paths, nil
}

// getKptFilesPaths returns all the paths to Kptfiles relative to the root package
func getKptFilesPaths(nodes []*yaml.RNode) (sets.String, error) {
	paths := sets.String{}
	for _, node := range nodes {
		meta, err := node.GetMeta()
		if err != nil {
			return nil, err
		}
		path := meta.Annotations[kioutil.PathAnnotation]
		if filepath.Base(path) == v1.KptFileName {
			paths.Insert(meta.Annotations[kioutil.PathAnnotation])
		}
	}
	return paths, nil
}

// filesInPackage returns all the file paths which belong to the input pkgPath
// this doesn't include files in subpackages
func filesInPackage(pkgPath string, resourcesPaths, kptFilePaths sets.String) sets.String {
	res := sets.String{}
	for _, resourcePath := range resourcesPaths.List() {
		dirPath := filepath.Dir(resourcePath)
		for {
			// check if the input pkgPath is the immediate parent package for the resource
			kfPath := filepath.Join(dirPath, v1.KptFileName)
			if kptFilePaths.Has(kfPath) {
				if dirPath == pkgPath {
					// this means the dirPath has a Kptfile and is a package
					// and dirPath is the target package for which we are searching the resource paths
					res.Insert(resourcePath)
				}
				break
			}
			if dirPath == "" || dirPath == "." {
				break
			}
			// keep searching the parent directory for Kptfile
			dirPath = filepath.Dir(dirPath)
		}
	}
	return res
}

// FunctionsInPkg gets the v1 functions list for functions in package
// nodes is list of input nodes which are sorted according to the package depth
// i is the index of the Kptfile of the package, all the files till i hits next Kptfile
// are the files of the package
// pkgPath is the package path relative to the root package directory
func (s *Fix) FunctionsInPkg(nodes []*yaml.RNode, pkgPath string) []v1.Function {
	var res []v1.Function
	for _, node := range nodes {
		meta, err := node.GetMeta()
		if err != nil {
			return res
		}
		nonKfPkgPaths := s.pkgPathToPkgFilePaths[pkgPath]
		if !nonKfPkgPaths.Has(meta.Annotations[kioutil.PathAnnotation]) {
			continue
		}
		fnSpec := runtimeutil.GetFunctionSpec(node)
		if fnSpec != nil {
			// in order to make sure there is only one function config per file,
			// rename the file with the name(and hash(name)) of the function config resource
			fnFileName := fmt.Sprintf("%s.yaml", node.GetName())
			// in v1, fn-config must be present in the package directory
			// so configPath must be just the file name
			fnFilePath := filepath.Join(pkgPath, fnFileName)
			res = append(res, v1.Function{
				Image:      fnSpec.Container.Image,
				ConfigPath: fnFileName,
			})
			// move the fn-config to the top level directory of the package
			meta.Annotations[kioutil.PathAnnotation] = fnFilePath
			delete(meta.Annotations, runtimeutil.FunctionAnnotationKey)
			delete(meta.Annotations, "config.k8s.io/function")
			err = node.SetAnnotations(meta.Annotations)
			if err != nil {
				return res
			}
		}
	}
	return res
}

// FixKptfile migrates the input Kptfile node from v1alpha1 or v1alpha2 to v1
func (s *Fix) FixKptfile(node *yaml.RNode, functions []v1.Function) (*yaml.RNode, error) {
	var err error
	meta, err := node.GetMeta()
	if err != nil {
		return node, err
	}

	settersConfigFilePath := filepath.Join(filepath.Dir(meta.Annotations[kioutil.PathAnnotation]), SettersConfigFileName)

	// v1alpha2 to v1 migration
	if strings.Contains(meta.APIVersion, "v1alpha2") {
		node.SetApiVersion(v1.KptFileAPIVersion)
		s.Results = append(s.Results, &Result{
			FilePath: meta.Annotations[kioutil.PathAnnotation],
			Message:  fmt.Sprintf("Updated apiVersion to %s", v1.KptFileAPIVersion),
		})

		kf, err := v1.ReadFile(node)
		if err != nil {
			return node, err
		}

		// apply-setters function input should be moved to configPath option as it
		// is the best practice
		if kf.Pipeline != nil {
			for i, fn := range kf.Pipeline.Mutators {
				if strings.Contains(fn.Image, "ghcr.io/kptdev/krm-functions-catalog/apply-setters:") && len(fn.ConfigMap) > 0 {
					settersConfig, err := SettersNodeFromSetters(kf.Pipeline.Mutators[0].ConfigMap, settersConfigFilePath)
					if err != nil {
						return node, err
					}
					s.settersConfigs = append(s.settersConfigs, settersConfig)
					kf.Pipeline.Mutators[i].ConfigMap = nil
					kf.Pipeline.Mutators[i].ConfigPath = SettersConfigFileName
					s.Results = append(s.Results, &Result{
						FilePath: settersConfigFilePath,
						Message:  `Moved setters from configMap to configPath`,
					})
				}
			}
		}

		// convert updated kf to yaml node
		b, err := yaml.Marshal(kf)
		if err != nil {
			return node, err
		}
		kNode, err := yaml.Parse(string(b))
		if err != nil {
			return node, err
		}
		err = kNode.SetAnnotations(meta.Annotations)
		return kNode, err
	}

	// return if the package with this Kptfile is already fixed
	if meta.APIVersion == v1.KptFileAPIVersion {
		s.Results = append(s.Results, &Result{
			FilePath: meta.Annotations[kioutil.PathAnnotation],
			Message:  fmt.Sprintf("This package is already fixed as it is on latest apiVersion %s", v1.KptFileAPIVersion),
		})
		return node, nil
	}

	// v1alpha1 to v1 migration
	kfOld, err := v1alpha1.ReadFile(node)
	if err != nil {
		return node, err
	}

	kfNew := v1.KptFile{ResourceMeta: meta}
	kfNew.APIVersion = v1.KptFileAPIVersion

	// convert packageMetadata in v1alpha1 Kptfile to v1 info
	if kfOld.PackageMeta != nil {
		emails := []string{kfOld.PackageMeta.Email}
		if kfOld.PackageMeta.Email == "" {
			emails = nil
		}
		kfNew.Info = &v1.PackageInfo{
			Site:        kfOld.PackageMeta.URL,
			Emails:      emails,
			License:     kfOld.PackageMeta.License,
			Description: kfOld.PackageMeta.ShortDescription,
			Keywords:    kfOld.PackageMeta.Tags,
			Man:         kfOld.PackageMeta.Man,
		}
		s.Results = append(s.Results, &Result{
			FilePath: meta.Annotations[kioutil.PathAnnotation],
			Message:  `Transformed "packageMetadata" to "info"`,
		})
	}

	// convert upstream section
	if kfOld.Upstream != nil {
		kfNew.Upstream = &v1.Upstream{
			Type: v1.OriginType(kfOld.Upstream.Type),
			Git: &v1.Git{
				Repo:      kfOld.Upstream.Git.Repo,
				Directory: kfOld.Upstream.Git.Directory,
				Ref:       kfOld.Upstream.Git.Ref,
			},
			UpdateStrategy: v1.ResourceMerge,
		}

		kfNew.UpstreamLock = &v1.UpstreamLock{
			Type: v1.OriginType(kfOld.Upstream.Type),
			Git: &v1.GitLock{
				Repo:      kfOld.Upstream.Git.Repo,
				Directory: kfOld.Upstream.Git.Directory,
				Ref:       kfOld.Upstream.Git.Ref,
				Commit:    kfOld.Upstream.Git.Commit,
			},
		}

		s.Results = append(s.Results, &Result{
			FilePath: meta.Annotations[kioutil.PathAnnotation],
			Message:  `Transformed "upstream" to "upstream" and "upstreamLock"`,
		})
	}

	if err != nil {
		return node, err
	}
	setters, err := listSetters(node)
	if err != nil {
		return node, err
	}

	pl := &v1.Pipeline{}
	kfNew.Pipeline = pl
	for _, fn := range functions {
		s.Results = append(s.Results, &Result{
			FilePath: meta.Annotations[kioutil.PathAnnotation],
			Message:  fmt.Sprintf(`Added %q to mutators list, please move it to validators section if it is a validator function`, fn.Image),
		})
	}

	if len(setters) > 0 {
		fn := v1.Function{
			// (TODO) If the fix fn is to remain, it should use a later version of apply-setters
			Image:      "ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.2",
			ConfigPath: SettersConfigFileName,
		}
		settersConfig, err := SettersNodeFromSetters(setters, settersConfigFilePath)
		if err != nil {
			return node, err
		}
		s.settersConfigs = append(s.settersConfigs, settersConfig)
		pl.Mutators = append(pl.Mutators, fn)
		s.Results = append(s.Results, &Result{
			FilePath: meta.Annotations[kioutil.PathAnnotation],
			Message:  `Transformed "openAPI" definitions to "apply-setters" function`,
		})
	}
	pl.Mutators = append(pl.Mutators, functions...)

	// convert inventory section
	if kfOld.Inventory != nil {
		kfNew.Inventory = &v1.Inventory{
			Namespace:   kfOld.Inventory.Namespace,
			Name:        kfOld.Inventory.Name,
			InventoryID: kfOld.Inventory.InventoryID,
			Labels:      kfOld.Inventory.Labels,
			Annotations: kfOld.Inventory.Annotations,
		}
	}

	// convert kfNew to yaml node
	b, err := yaml.Marshal(kfNew)
	if err != nil {
		return node, err
	}
	kNode, err := yaml.Parse(string(b))
	if err != nil {
		return node, err
	}
	err = kNode.SetAnnotations(meta.Annotations)
	return kNode, err
}

// SettersNodeFromSetters returns the ConfigMap node with input setters in data field
func SettersNodeFromSetters(setters map[string]string, path string) (*yaml.RNode, error) {
	configNode, err := yaml.Parse(fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: setters
  annotations:
    %s: %s
    %s: "0"
`, kioutil.PathAnnotation, path, kioutil.IndexAnnotation))
	if err != nil {
		return nil, err
	}

	var keys []string
	for k := range setters {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v, err := standardizeArrayValue(setters[k])
		if err != nil {
			return nil, err
		}
		vNode := yaml.NewScalarRNode(v)
		if v == "" {
			vNode.YNode().Style = yaml.DoubleQuotedStyle
		}
		err = configNode.PipeE(
			yaml.LookupCreate(yaml.ScalarNode, "data", k),
			yaml.FieldSetter{Value: vNode, OverrideStyle: true})
		if err != nil {
			return nil, err
		}
	}
	// format the created resource
	_, err = filters.FormatFilter{UseSchema: true}.Filter([]*yaml.RNode{configNode})
	return configNode, err
}

// standardizeArrayValue returns the folded style array node string
// e.g. for input value [foo, bar], it returns
// - foo
// - bar
func standardizeArrayValue(val string) (string, error) {
	if !strings.HasPrefix(val, "[") {
		// the value is either standardized, or is not a sequence node value
		return val, nil
	}
	vNode, err := yaml.Parse(val)
	if err != nil {
		return val, fmt.Errorf("failed to parse the array node value %q with error %q", val, err.Error())
	}
	if vNode.YNode().Kind == yaml.SequenceNode {
		// standardize array values to folded style
		vNode.YNode().Style = yaml.FoldedStyle
		val, err = vNode.String()
		if err != nil {
			return val, fmt.Errorf("failed to serialize the array node value %q with error %q", val, err.Error())
		}
	}
	return val, nil
}

// getPkgPathToSettersSchema returns the, map of pkgPath to the setters schema in Kptfile
// of the package
func getPkgPathToSettersSchema(nodes []*yaml.RNode) (map[string]*spec.Schema, error) {
	res := make(map[string]*spec.Schema)
	for _, node := range nodes {
		meta, err := node.GetMeta()
		if err != nil {
			return nil, err
		}
		if meta.Kind == v1.KptFileName {
			// convert OpenAPI section in v1alpha1 Kptfile to apply-setters
			schema, err := schemaUsingField(node, openapi.SupplementaryOpenAPIFieldName)
			if err != nil {
				return nil, err
			}
			res[filepath.Dir(meta.Annotations[kioutil.PathAnnotation])] = schema
		}
	}
	return res, nil
}

// visitMapping visits mapping node to convert the comments for array setters
func (s *Fix) visitMapping(object *yaml.RNode) error {
	return object.VisitFields(func(node *yaml.MapNode) error {
		if node.IsNilOrEmpty() {
			return nil
		}
		if node.Value.YNode().Kind != yaml.SequenceNode {
			// return if it is not a sequence node
			return nil
		}
		oldCommentPrefixes := []string{`# {"$kpt-set":"`, `# {"$ref":"#/definitions/io.k8s.cli.`}
		comment := node.Key.YNode().LineComment
		for _, cp := range oldCommentPrefixes {
			// # {"$kpt-set":"foo"} must be converted to # kpt-set: ${foo}
			// # {"$ref":"#/definitions/io.k8s.cli.list"} must be converted to # kpt-set: ${list}
			if strings.Contains(comment, cp) {
				comment := strings.TrimPrefix(comment, cp)
				comment = strings.TrimSuffix(comment, `"}`)
				node.Key.YNode().LineComment = fmt.Sprintf("kpt-set: ${%s}", comment)
			}
		}
		return nil
	})
}

// visitScalar visits scalar nodes and converts the comments to v1 format
func (s *Fix) visitScalar(object *yaml.RNode, setterSchema *openapi.ResourceSchema) error {
	ext, err := getExtFromComment(setterSchema)
	if err != nil {
		return err
	}
	if ext == nil {
		return nil
	}

	ok, err := s.fixSetter(object, ext)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	_, err = s.fixSubst(object, ext)
	if err != nil {
		return err
	}
	return nil
}

// fixSetter converts the setter comment to v1 format
func (s *Fix) fixSetter(field *yaml.RNode, ext *setters2.CliExtension) (bool, error) {
	// check full setter
	if ext == nil || ext.Setter == nil {
		return false, nil
	}

	field.YNode().LineComment = fmt.Sprintf("kpt-set: ${%s}", ext.Setter.Name)
	return true, nil
}

// fixSubst converts the substitution comment to expanded setter comment pattern
func (s *Fix) fixSubst(field *yaml.RNode, ext *setters2.CliExtension) (bool, error) {
	if ext.Substitution == nil {
		return false, nil
	}

	// track the visited nodes to detect cycles in nested substitutions
	visited := sets.String{}

	res, err := s.substituteUtil(ext, visited)
	if err != nil {
		return false, err
	}

	field.YNode().LineComment = fmt.Sprintf("kpt-set: %s", res)

	return true, nil
}

// substituteUtil recursively parses nested substitutions in ext and sets the setter value
// returns error if cyclic substitution is detected or any other unexpected errors
func (s *Fix) substituteUtil(ext *setters2.CliExtension, visited sets.String) (string, error) {
	// check if the substitution has already been visited and throw error as cycles
	// are not allowed in nested substitutions
	if visited.Has(ext.Substitution.Name) {
		return "", errors.Errorf(
			"cyclic substitution detected with name %s", ext.Substitution.Name)
	}

	visited.Insert(ext.Substitution.Name)
	pattern := ext.Substitution.Pattern

	// substitute each setter into the pattern to get the new value
	// if substitution references to another substitution, recursively
	// process the nested substitutions to replace the pattern with setter values
	for _, v := range ext.Substitution.Values {
		if v.Ref == "" {
			return "", errors.Errorf(
				"missing reference on substitution %s", ext.Substitution.Name)
		}
		ref, err := spec.NewRef(v.Ref)
		if err != nil {
			return "", errors.Wrap(err)
		}
		def, err := openapi.Resolve(&ref, s.settersSchema) // resolve the def to its openAPI def
		if err != nil {
			return "", errors.Wrap(err)
		}
		defExt, err := setters2.GetExtFromSchema(def) // parse the extension out of the openAPI
		if err != nil {
			return "", errors.Wrap(err)
		}

		if defExt.Substitution != nil {
			// parse recursively if it reference is substitution
			substVal, err := s.substituteUtil(defExt, visited)
			if err != nil {
				return "", err
			}
			pattern = strings.ReplaceAll(pattern, v.Marker, substVal)
			continue
		}

		if val, found := defExt.Setter.EnumValues[defExt.Setter.Value]; found {
			// the setter has an enum-map. we should replace the marker with the
			// enum value looked up from the map rather than the enum key
			pattern = strings.ReplaceAll(pattern, v.Marker, val)
		} else {
			pattern = strings.ReplaceAll(pattern, v.Marker, fmt.Sprintf("${%s}", defExt.Setter.Name))
		}
	}
	return pattern, nil
}

// listSetters extracts the setters information from the input Kptfile node
func listSetters(object *yaml.RNode) (map[string]string, error) {
	setters := make(map[string]string)
	// read the OpenAPI definitions in Kptfile
	def, err := object.Pipe(yaml.LookupCreate(yaml.MappingNode, "openAPI", "definitions"))
	if err != nil {
		return nil, err
	}
	if yaml.IsMissingOrNull(def) {
		return nil, nil
	}

	// iterate over definitions -- find those that are setters
	err = def.VisitFields(func(node *yaml.MapNode) error {
		setter := setters2.SetterDefinition{}

		// the definition key -- contains the setter name
		key := node.Key.YNode().Value

		if !strings.HasPrefix(key, fieldmeta.SetterDefinitionPrefix) {
			// not a setter -- doesn't have the right prefix
			return nil
		}

		setterNode, err := node.Value.Pipe(yaml.Lookup(setters2.K8sCliExtensionKey, "setter"))
		if err != nil {
			return err
		}
		if yaml.IsMissingOrNull(setterNode) {
			// has the setter prefix, but missing the setter extension
			return errors.Errorf("missing x-k8s-cli.setter for %s", key)
		}

		// unmarshal the yaml for the setter extension into the definition struct
		b, err := setterNode.String()
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal([]byte(b), &setter); err != nil {
			return err
		}

		// the description is not part of the extension, and should be pulled out
		// separately from the extension values.
		description := node.Value.Field("description")
		if description != nil {
			setter.Description = description.Value.YNode().Value
		}
		if len(setter.ListValues) > 0 {
			var vals string
			for _, val := range setter.ListValues {
				vals = vals + ", " + val
			}
			vals = strings.TrimPrefix(vals, ", ")
			setters[setter.Name] = fmt.Sprintf("[%s]", vals)
		} else {
			setters[setter.Name] = setter.Value
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return setters, nil
}

// getExtFromComment returns the cliExtension openAPI extension if it is present as
// a comment on the field.
func getExtFromComment(schema *openapi.ResourceSchema) (*setters2.CliExtension, error) {
	if schema == nil {
		return nil, nil
	}

	// get the cli extension from the openapi (contains setter information)
	ext, err := setters2.GetExtFromSchema(schema.Schema)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return ext, nil
}

// schemaUsingField returns the schema object for the openAPI section in input Kptfile object node
func schemaUsingField(object *yaml.RNode, field string) (*spec.Schema, error) {
	if field != "" {
		// get the field containing the openAPI
		m := object.Field(field)
		if m.IsNilOrEmpty() {
			// doesn't contain openAPI definitions
			return nil, nil
		}
		object = m.Value
	}

	oAPI, err := object.String()
	if err != nil {
		return nil, err
	}

	// convert the yaml openAPI to a JSON string by unmarshalling it to an
	// interface{} and the marshalling it to a string
	var o interface{}
	err = yaml.Unmarshal([]byte(oAPI), &o)
	if err != nil {
		return nil, err
	}
	j, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	var sc spec.Schema
	err = sc.UnmarshalJSON(j)
	if err != nil {
		return nil, err
	}

	return &sc, nil
}
