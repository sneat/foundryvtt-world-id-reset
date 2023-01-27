package parser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	appFs = afero.NewOsFs()
)

// Service for parsing the database files
type Service struct {
	// Options for the parser
	Options Options

	logger *logrus.Logger

	hasWorldData  bool
	hasModuleData bool
}

// Run parses the database files and converts the IDs
func Run(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.PersistentFlags().GetBool("verbose")
	modulePath, err := cmd.PersistentFlags().GetString("path")
	if err != nil {
		return err
	}

	s := &Service{
		Options: NewOptions(
			Verbose(verbose),
			Path(modulePath),
		),
		logger: logrus.New(),
	}

	if verbose {
		s.logger.SetLevel(logrus.DebugLevel)
	}

	s.logger.WithFields(logrus.Fields{
		"options": fmt.Sprintf("%+v", s.Options),
	}).Debug("Running with provided configuration.")

	if err = s.Validate(); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	s.logger.Info("Ensure that your Foundry VTT world is NOT running. You will likely corrupt your database files if they are actively being referenced by Foundry VTT. Ctrl-C this program and stop Foundry VTT if need be.")
	s.logger.Info("Press 'Enter' to continue...")
	if _, err = bufio.NewReader(os.Stdin).ReadBytes('\n'); err != nil {
		return err
	}

	var files []string

	if s.hasWorldData {
		worldFiles, err := afero.Glob(appFs, filepath.Join(s.Options.Path, "data", "*.db"))
		if err != nil {
			return fmt.Errorf("error finding database files: %w", err)
		}
		files = append(files, worldFiles...)
	}

	if s.hasModuleData {
		moduleFiles, err := afero.Glob(appFs, filepath.Join(s.Options.Path, "packs", "*.db"))
		if err != nil {
			return fmt.Errorf("error finding database files: %w", err)
		}
		files = append(files, moduleFiles...)
	}

	if len(files) == 0 {
		s.logger.WithField("path", filepath.Join(s.Options.Path, "data", "*.db")).Error("No database files found.")
		return errors.New("no database files found")
	}

	s.logger.WithField("files", files).Debugf("Parsing %d database files.", len(files))

	var idMap = make(map[string]string)
	var sceneIDMap = make(map[string]string) // Scenes need their thumbnails updated too
	var didError bool

	for _, file := range files {
		filename := filepath.Base(file)
		fileLogger := s.logger.WithField("file", filename)
		docs, err := s.ParseFile(file)
		if err != nil {
			return fmt.Errorf("error parsing database file: %w", err)
		}

		fileLogger.Debugf("Parsed %d documents.", len(docs))
		var output strings.Builder
		if len(docs) > 0 {
			output.WriteString(fmt.Sprintf("%s ID Map:\n", filename))
		}

		for _, doc := range docs {
			docLogger := fileLogger.WithField("name", doc.Name)
			if _, ok := idMap[doc.OldID]; ok {
				docLogger.WithField("old_id", doc.OldID).Error("Duplicate ID found.")
				didError = true
			}

			if _, ok := idMap[doc.NewID]; ok {
				docLogger.WithField("new_id", doc.NewID).Error("Duplicate ID found.")
				didError = true
			}

			idMap[doc.OldID] = doc.NewID

			if filename == "scenes.db" {
				sceneIDMap[doc.OldID] = doc.NewID
			}

			if len(docs) > 0 {
				output.WriteString(fmt.Sprintf("%s: %s -> %s\n", doc.Name, doc.OldID, doc.NewID))
			}
		}

		if len(docs) > 0 {
			fileLogger.Debug(output.String())
		}
	}

	if didError {
		return errors.New("duplicate IDs found")
	}

	s.logger.Info("ID Mapping has been generated. Run this program with the --verbose flag to see the mapping.")
	s.logger.Info("Ctrl-C this program if you do not want to proceed with the update.")
	s.logger.Info("Press 'Enter' to continue...")
	if _, err = bufio.NewReader(os.Stdin).ReadBytes('\n'); err != nil {
		return err
	}

	s.logger.Info("Updating database files.")

	for _, file := range files {
		if err = s.UpdateFile(file, idMap); err != nil {
			return fmt.Errorf("error updating database file: %w", err)
		}
	}

	for oldID, newID := range sceneIDMap {
		if oldID == newID {
			continue
		}

		path := filepath.Join(s.Options.Path, "scenes", "thumbs")
		oldPath := filepath.Join(path, oldID+".png")
		newPath := filepath.Join(path, newID+".png")

		s.logger.WithFields(logrus.Fields{
			"old_path": filepath.Base(oldPath),
			"new_path": filepath.Base(newPath),
		}).Debug("Renaming scene thumbnail.")

		if ok, err := afero.Exists(appFs, oldPath); err != nil || !ok {
			if err != nil {
				return fmt.Errorf("error checking for thumbnail file: %w", err)
			}

			if !ok {
				return fmt.Errorf("thumbnail file not found: %s", oldPath)
			}
		}

		if err = appFs.Rename(oldPath, newPath); err != nil {
			return fmt.Errorf("error renaming thumbnail file: %w", err)
		}
	}

	return nil
}

// ParseFile parses a database file
func (s *Service) ParseFile(file string) ([]Document, error) {
	fileLogger := s.logger.WithField("file", filepath.Base(file))
	fileLogger.Debug("Parsing database file.")

	// Read the file line by line and create Document structs
	fs, err := appFs.Open(file)
	if err != nil {
		return nil, fmt.Errorf("error opening database file: %w", err)
	}
	defer fs.Close()

	var docs []Document

	scanner := bufio.NewScanner(fs)
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	for scanner.Scan() {
		doc, err := ParseDocument(scanner.Bytes())
		if err != nil {
			fileLogger.WithError(err).Error("Error parsing document.")
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

// UpdateFile updates a database file with the new IDs
func (s *Service) UpdateFile(file string, idMap map[string]string) error {
	fileLogger := s.logger.WithField("file", filepath.Base(file))
	fileLogger.Debug("Updating database file.")

	// Replace the old IDs with the new IDs within the entire file
	input, err := afero.ReadFile(appFs, file)
	if err != nil {
		return fmt.Errorf("error opening database file: %w", err)
	}

	output := input
	for oldID, newID := range idMap {
		output = bytes.ReplaceAll(output, []byte(oldID), []byte(newID))
	}

	if err = afero.WriteFile(appFs, file, output, 0666); err != nil {
		return fmt.Errorf("error writing database file: %w", err)
	}

	return nil
}

// Validate the options
func (s *Service) Validate() error {
	worldExists, err := afero.Exists(appFs, filepath.Join(s.Options.Path, "world.json"))
	if err != nil {
		return fmt.Errorf("error checking for world.json: %w", err)
	}

	moduleExists, err := afero.Exists(appFs, filepath.Join(s.Options.Path, "module.json"))
	if err != nil {
		return fmt.Errorf("error checking for module.json: %w", err)
	}

	if !worldExists && !moduleExists {
		return errors.New("world.json or module.json not found, check this is a Foundry VTT world or module directory")
	}

	if worldExists {
		s.hasWorldData = true
		if ok, err := afero.DirExists(appFs, filepath.Join(s.Options.Path, "data")); err != nil || !ok {
			if err != nil {
				return fmt.Errorf("error checking for world database directory: %w", err)
			}
			return errors.New("world database directory not found, check this is a Foundry VTT world directory")
		}
	}

	if moduleExists {
		s.hasModuleData = true
		if ok, err := afero.DirExists(appFs, filepath.Join(s.Options.Path, "packs")); err != nil || !ok {
			if err != nil {
				return fmt.Errorf("error checking for module database directory: %w", err)
			}
			return errors.New("module database directory not found, check this is a Foundry VTT module directory")
		}
	}

	return nil
}

// NewRandomID generates a new random ID.
//   - 16 characters
func NewRandomID() string {
	rand.Seed(time.Now().UnixNano())

	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 16)

	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}

	return string(result)
}
